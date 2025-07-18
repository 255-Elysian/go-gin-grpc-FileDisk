package service

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"grpc-todolist-disk/app/files/internal/repository/dao"
	pb "grpc-todolist-disk/idl/pb/files"
	"grpc-todolist-disk/utils/e"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var FilesSrvIns *FilesSrv
var FilesSrvOnce sync.Once

type FilesSrv struct {
	pb.UnimplementedFilesServiceServer
}

func GetFilesSrv() *FilesSrv {
	FilesSrvOnce.Do(func() {
		FilesSrvIns = &FilesSrv{}
	})
	return FilesSrvIns
}

// FileUpload 文件上传（表单）
func (*FilesSrv) FileUpload(ctx context.Context, req *pb.FileUploadRequest) (resp *pb.FileUploadResponse, err error) {
	resp = new(pb.FileUploadResponse)
	resp.Code = e.SUCCESS
	resp.ObjectUrl = filepath.Join("stores/uploaded_files", req.ObjectName)
	file, err := dao.NewFilesDao().CreateFile(req)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(int(resp.Code))
		return
	}
	resp.FileID = uint64(file.ID)
	resp.Msg = e.GetMsg(int(resp.Code))
	return
}

// BigFileUpload 文件上传（流上传）
func (*FilesSrv) BigFileUpload(stream pb.FilesService_BigFileUploadServer) error {
	var (
		firstReq   *pb.BigFileUploadRequest
		objectPath string
		totalSize  int64
		out        *os.File
	)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break // 接收完毕
		}
		if err != nil {
			return stream.SendAndClose(&pb.BigFileUploadResponse{
				Code: e.ERROR,
				Msg:  "接收上传流失败: " + err.Error(),
			})
		}

		if firstReq == nil {
			firstReq = req

			// 生成路径
			objectPath = filepath.Join("stores/uploaded_files", req.ObjectName)
			if err := os.MkdirAll(filepath.Dir(objectPath), os.ModePerm); err != nil {
				return stream.SendAndClose(&pb.BigFileUploadResponse{
					Code: e.ERROR,
					Msg:  "创建目录失败: " + err.Error(),
				})
			}

			out, err = os.OpenFile(objectPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return stream.SendAndClose(&pb.BigFileUploadResponse{
					Code: e.ERROR,
					Msg:  "创建文件失败: " + err.Error(),
				})
			}
			defer out.Close()
		}

		// 写入内容
		n, err := out.Write(req.Content)
		if err != nil {
			return stream.SendAndClose(&pb.BigFileUploadResponse{
				Code: e.ERROR,
				Msg:  "文件写入失败: " + err.Error(),
			})
		}
		totalSize += int64(n)

		if req.IsLast {
			firstReq.FileSize = totalSize
			break
		}
	}

	// 防御性检查：防止客户端没有发送任何分片就关闭了流
	if firstReq == nil {
		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code: e.ERROR,
			Msg:  "上传内容为空",
		})
	}
	firstReq.FileSize = totalSize

	// 数据库保存记录
	//log.Println("收到上传总大小：", totalSize)
	firstReq.FileSize = totalSize
	file, err := dao.NewFilesDao().CreateBigFile(firstReq)
	if err != nil {
		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code: e.ERROR,
			Msg:  e.GetMsg(e.ERROR),
		})
	}

	return stream.SendAndClose(&pb.BigFileUploadResponse{
		Code:      e.SUCCESS,
		Msg:       e.GetMsg(e.SUCCESS),
		FileID:    uint64(file.ID),
		ObjectUrl: objectPath,
	})
}

// FileList 获取用户文件列表
func (*FilesSrv) FileList(ctx context.Context, req *pb.FileListRequest) (resp *pb.FileListResponse, err error) {
	resp = new(pb.FileListResponse)
	resp.Code = e.SUCCESS
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	files, total, err := dao.NewFilesDao().ListFiles(req)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(int(resp.Code))
		return
	}
	resp.Total = total
	for _, file := range files {
		resp.Files = append(resp.Files, &pb.FileModel{
			FileID:     uint64(file.ID),
			UserID:     uint64(file.UserID),
			FileName:   file.FileName,
			FileSize:   file.FileSize,
			Bucket:     file.Bucket,
			ObjectName: file.ObjectName,
		})
	}
	resp.Msg = e.GetMsg(int(resp.Code))
	return
}

// FileDelete 删除用户文件
func (*FilesSrv) FileDelete(ctx context.Context, req *pb.FileDeleteRequest) (resp *pb.FileCommonResponse, err error) {
	resp = new(pb.FileCommonResponse)
	resp.Code = e.SUCCESS

	file, err := dao.NewFilesDao().GetFileByUIDAndFID(uint(req.UserID), uint(req.FileID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.Code = e.ERROR
			resp.Msg = "文件不存在"
			return resp, nil
		}
		resp.Code = e.ERROR
		resp.Msg = "查询文件信息失败"
		return
	}
	objectPath := filepath.Join("stores/uploaded_files", file.ObjectName)
	if err = os.Remove(objectPath); err != nil && !os.IsNotExist(err) {
		resp.Code = e.ERROR
		resp.Msg = "删除本地文件失败: " + err.Error()
		return
	}

	err = dao.NewFilesDao().DeleteFile(req)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(int(resp.Code))
		return
	}
	resp.Msg = e.GetMsg(e.SUCCESS)

	zap.L().Info("Delete file", zap.Uint64("user_id", req.UserID), zap.Uint64("file_id", req.FileID), zap.String("path", objectPath))
	return
}

// FileDownload 下载文件
func (*FilesSrv) FileDownload(ctx context.Context, req *pb.FileDownloadRequest) (resp *pb.FileDownloadResponse, err error) {
	resp = new(pb.FileDownloadResponse)
	resp.Code = e.SUCCESS

	file, err := dao.NewFilesDao().GetFileByUIDAndFID(uint(req.UserID), uint(req.FileID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.Code = e.ERROR
			resp.Msg = "文件不存在"
			return resp, nil
		}
		resp.Code = e.ERROR
		resp.Msg = "查询文件信息失败"
		return
	}
	resp.DownloadUrl = filepath.Join("stores/uploaded_files", file.ObjectName)
	resp.Filename = file.FileName
	resp.Msg = e.GetMsg(int(resp.Code))
	return
}
