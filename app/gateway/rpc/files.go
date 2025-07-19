package rpc

import (
	"context"
	"errors"
	"fmt"
	"grpc-todolist-disk/app/gateway/utils"
	pb "grpc-todolist-disk/idl/pb/files"
	"grpc-todolist-disk/utils/e"
	"io"
	"time"
)

func FileUpload(ctx context.Context, req *pb.FileUploadRequest) (resp *pb.FileUploadResponse, err error) {
	resp, err = FilesClient.FileUpload(ctx, req)

	if err != nil {
		return
	}

	if resp.Code != e.SUCCESS {
		err = errors.New(resp.Msg)
		return
	}

	return
}

type UploadMeta struct {
	UserID   uint64
	FileName string
	FileSize int64
}

// BigFileUpload 分片上传大文件
func BigFileUpload(ctx context.Context, reader io.Reader, meta *UploadMeta) (*pb.BigFileUploadResponse, error) {
	stream, err := FilesClient.BigFileUpload(ctx)
	if err != nil {
		return nil, fmt.Errorf("初始化上传流失败: %w", err)
	}

	objectName := fmt.Sprintf("%d/%d_%s", meta.UserID, time.Now().UnixMilli(), utils.Clean(meta.FileName))

	const chunkSize = 1 << 20 // 1MB
	buf := make([]byte, chunkSize)
	totalSize := int64(0)
	first := true

	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("读取文件数据失败: %w", err)
		}
		if n == 0 {
			break
		}

		totalSize += int64(n)

		req := &pb.BigFileUploadRequest{
			UserID:     meta.UserID,
			Filename:   meta.FileName,
			FileSize:   0, // 只在最后发送
			ObjectName: objectName,
			Content:    buf[:n],
			IsLast:     false,
		}

		if first {
			first = false
		}

		if err == io.EOF {
			req.IsLast = true
			req.FileSize = totalSize
		}

		if err := stream.Send(req); err != nil {
			return nil, fmt.Errorf("发送分片失败: %w", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("关闭上传流失败: %w", err)
	}

	return res, nil
}

func FileList(ctx context.Context, req *pb.FileListRequest) (resp *pb.FileListResponse, err error) {
	resp, err = FilesClient.FileList(ctx, req)
	if err != nil {
		return
	}

	if resp.Code != e.SUCCESS {
		err = errors.New(resp.Msg)
		return
	}

	return
}

func FileDelete(ctx context.Context, req *pb.FileDeleteRequest) (resp *pb.FileCommonResponse, err error) {
	resp, err = FilesClient.FileDelete(ctx, req)
	if err != nil {
		return
	}

	if resp.Code != e.SUCCESS {
		err = errors.New(resp.Msg)
		return
	}
	return
}

func FileDownload(ctx context.Context, req *pb.FileDownloadRequest) (resp *pb.FileDownloadResponse, err error) {
	resp, err = FilesClient.FileDownload(ctx, req)
	if err != nil {
		return
	}

	if resp.Code != e.SUCCESS {
		err = errors.New(resp.Msg)
		return
	}
	return
}

func CheckFileExists(ctx context.Context, req *pb.CheckFileRequest) (resp *pb.CheckFileResponse, err error) {
	resp, err = FilesClient.CheckFileExists(ctx, req)
	if err != nil {
		return
	}

	return
}
