package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"grpc-todolist-disk/app/files/dao"
	"grpc-todolist-disk/app/files/internal/repository/utils"
	pb "grpc-todolist-disk/idl/pb/files"
	"grpc-todolist-disk/utils/e"
	"io"
	"log"
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
		hashes     = sha256.New() // 创建 Hash 实例
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

			// 写入临时路径
			objectPath = filepath.Join("stores/uploaded_temp", req.ObjectName)
			if err = os.MkdirAll(filepath.Dir(objectPath), os.ModePerm); err != nil {
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
			//defer out.Close()
		}

		// 写入 Hash 内容
		if _, err = hashes.Write(req.Content); err != nil {
			return stream.SendAndClose(&pb.BigFileUploadResponse{
				Code: e.ERROR,
				Msg:  "计算 Hash 错误: " + err.Error(),
			})
		}

		// 写入磁盘
		n, err := out.Write(req.Content)
		if err != nil {
			return stream.SendAndClose(&pb.BigFileUploadResponse{
				Code: e.ERROR,
				Msg:  "文件写入失败: " + err.Error(),
			})
		}
		totalSize += int64(n)

		if req.IsLast {
			// 最后一块后立即关闭文件
			if err := out.Close(); err != nil {
				return stream.SendAndClose(&pb.BigFileUploadResponse{
					Code: e.ERROR,
					Msg:  "文件关闭失败: " + err.Error(),
				})
			}
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

	// 计算最终 Hash 值
	fileHash := hex.EncodeToString(hashes.Sum(nil))
	firstReq.FileHash = fileHash
	log.Println("FileHash:", fileHash)
	// 检查数据库是否已有相同文件
	exist, err := dao.NewFilesDao().FindByHash(&pb.CheckFileRequest{
		FileHash: firstReq.FileHash,
		UserID:   firstReq.UserID,
	})
	if err != nil {
		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code: e.ERROR,
			Msg:  "检查文件 Hash 失败: " + err.Error(),
		})
	}
	if exist != nil {
		utils.SafeRemove(objectPath) // 删除临时文件（忽略错误）
		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code:      e.SUCCESS,
			Msg:       "秒传成功，文件已存在",
			FileID:    uint64(exist.ID),
			ObjectUrl: filepath.Join("stores/uploaded_files", exist.ObjectName),
		})
	}

	// 将文件移到正式目录
	finalPath := filepath.Join("stores/uploaded_files", firstReq.ObjectName)
	if err = os.MkdirAll(filepath.Dir(finalPath), os.ModePerm); err != nil {
		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code: e.ERROR,
			Msg:  "创建目标目录失败: " + err.Error(),
		})
	}
	if err = os.Rename(objectPath, finalPath); err != nil {
		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code: e.ERROR,
			Msg:  "移动文件失败: " + err.Error(),
		})
	}

	// 数据库保存记录
	//log.Println("收到上传总大小：", totalSize)
	firstReq.FileSize = totalSize
	file, err := dao.NewFilesDao().CreateBigFile(firstReq)
	if err != nil {
		utils.SafeRemove(finalPath) // 删除已经移动过去的正式文件
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

// CheckFileExists 秒传哈希检测
func (*FilesSrv) CheckFileExists(ctx context.Context, req *pb.CheckFileRequest) (*pb.CheckFileResponse, error) {
	file, err := dao.NewFilesDao().FindByHash(req)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return &pb.CheckFileResponse{Exists: false}, nil
	}
	return &pb.CheckFileResponse{
		FileID:    uint64(file.ID),
		ObjectUrl: filepath.Join("stores/uploaded_files", file.ObjectName),
		Exists:    true,
	}, nil
}
