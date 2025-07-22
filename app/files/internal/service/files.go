package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"grpc-todolist-disk/app/files/dao"
	"grpc-todolist-disk/app/files/internal/repository/model"
	"grpc-todolist-disk/app/files/internal/repository/utils"
	pb "grpc-todolist-disk/idl/pb/files"
	"grpc-todolist-disk/utils/e"
	"grpc-todolist-disk/utils/qiniu"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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

// QiniuFileUpload 七牛云表单上传
func (*FilesSrv) QiniuFileUpload(ctx context.Context, req *pb.FileUploadRequest) (resp *pb.FileUploadResponse, err error) {
	resp = new(pb.FileUploadResponse)
	resp.Code = e.SUCCESS

	// 检查是否已存在相同文件（秒传）
	if req.FileHash != "" {
		// 先检查当前用户是否已有该文件（包括真实记录和秒传记录）
		userFile, err := dao.NewFilesDao().FindUserFileByOriginalHash(req.UserID, req.FileHash)
		if err != nil {
			resp.Code = e.ERROR
			resp.Msg = "检查用户文件失败: " + err.Error()
			return resp, nil
		}
		if userFile != nil {
			// 用户已有该文件（可能是真实记录或秒传记录），直接返回
			var objectUrl string
			if userFile.FileHash == req.FileHash {
				// 真实记录，直接使用其 ObjectName
				objectUrl = userFile.ObjectName
			} else {
				// 秒传记录，需要找到原始文件的 ObjectName
				originalFile, err := dao.NewFilesDao().FindGlobalByHash(req.FileHash)
				if err != nil {
					resp.Code = e.ERROR
					resp.Msg = "查找原始文件失败: " + err.Error()
					return resp, nil
				}
				objectUrl = originalFile.ObjectName
			}

			resp.FileID = uint64(userFile.ID)
			resp.ObjectUrl = objectUrl
			resp.Msg = "秒传成功，文件已存在"
			return resp, nil
		}

		// 检查全局是否存在相同哈希的文件
		globalFile, err := dao.NewFilesDao().FindGlobalByHash(req.FileHash)
		if err != nil {
			resp.Code = e.ERROR
			resp.Msg = "检查全局文件 Hash 失败: " + err.Error()
			return resp, nil
		}
		if globalFile != nil {
			// 全局存在相同文件，为当前用户创建新记录
			newUserFile, err := dao.NewFilesDao().CreateUserFileFromExisting(req.UserID, req.Filename, globalFile)
			if err != nil {
				resp.Code = e.ERROR
				resp.Msg = "创建用户文件记录失败: " + err.Error()
				return resp, nil
			}

			resp.FileID = uint64(newUserFile.ID)
			resp.ObjectUrl = globalFile.ObjectName
			resp.Msg = "秒传成功，文件已存在"
			return resp, nil
		}
	}

	// 如果没有 ObjectName，说明是秒传检查调用，不需要创建记录
	if req.ObjectName == "" {
		resp.Code = e.ERROR
		resp.Msg = "需要先上传文件到七牛云"
		return resp, nil
	}

	// 保存文件记录到数据库
	file, err := dao.NewFilesDao().CreateQiniuFile(req)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(int(resp.Code))
		return
	}

	resp.FileID = uint64(file.ID)
	resp.ObjectUrl = req.ObjectName // 返回七牛云的完整URL
	resp.Msg = e.GetMsg(int(resp.Code))
	return
}

// QiniuBigFileUpload 七牛云流式上传
func (*FilesSrv) QiniuBigFileUpload(stream pb.FilesService_QiniuBigFileUploadServer) error {
	var (
		firstReq  *pb.BigFileUploadRequest
		totalSize int64
		hashes    = sha256.New()
		chunks    [][]byte // 存储所有分片数据
	)

	// 接收所有分片数据
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return stream.SendAndClose(&pb.BigFileUploadResponse{
				Code: e.ERROR,
				Msg:  "接收上传流失败: " + err.Error(),
			})
		}

		if firstReq == nil {
			firstReq = req
		}

		// 计算 Hash
		if _, err = hashes.Write(req.Content); err != nil {
			return stream.SendAndClose(&pb.BigFileUploadResponse{
				Code: e.ERROR,
				Msg:  "计算 Hash 错误: " + err.Error(),
			})
		}

		// 存储分片数据
		chunks = append(chunks, req.Content)
		totalSize += int64(len(req.Content))

		if req.IsLast {
			break
		}
	}

	if firstReq == nil {
		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code: e.ERROR,
			Msg:  "上传内容为空",
		})
	}

	// 计算文件 Hash
	fileHash := hex.EncodeToString(hashes.Sum(nil))
	firstReq.FileHash = fileHash

	// 检查秒传
	// 先检查当前用户是否已有该文件（包括真实记录和秒传记录）
	userFile, err := dao.NewFilesDao().FindUserFileByOriginalHash(firstReq.UserID, firstReq.FileHash)
	if err != nil {
		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code: e.ERROR,
			Msg:  "检查用户文件失败: " + err.Error(),
		})
	}
	if userFile != nil {
		// 用户已有该文件（可能是真实记录或秒传记录），直接返回
		var objectUrl string
		if userFile.FileHash == firstReq.FileHash {
			// 真实记录，直接使用其 ObjectName
			objectUrl = userFile.ObjectName
		} else {
			// 秒传记录，需要找到原始文件的 ObjectName
			originalFile, err := dao.NewFilesDao().FindGlobalByHash(firstReq.FileHash)
			if err != nil {
				return stream.SendAndClose(&pb.BigFileUploadResponse{
					Code: e.ERROR,
					Msg:  "查找原始文件失败: " + err.Error(),
				})
			}
			objectUrl = originalFile.ObjectName
		}

		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code:      e.SUCCESS,
			Msg:       "秒传成功，文件已存在",
			FileID:    uint64(userFile.ID),
			ObjectUrl: objectUrl,
		})
	}

	// 检查全局是否存在相同哈希的文件
	globalFile, err := dao.NewFilesDao().FindGlobalByHash(firstReq.FileHash)
	if err != nil {
		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code: e.ERROR,
			Msg:  "检查全局文件 Hash 失败: " + err.Error(),
		})
	}
	if globalFile != nil {
		// 全局存在相同文件，为当前用户创建新记录
		newUserFile, err := dao.NewFilesDao().CreateUserFileFromExisting(firstReq.UserID, firstReq.Filename, globalFile)
		if err != nil {
			return stream.SendAndClose(&pb.BigFileUploadResponse{
				Code: e.ERROR,
				Msg:  "创建用户文件记录失败: " + err.Error(),
			})
		}

		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code:      e.SUCCESS,
			Msg:       "秒传成功，文件已存在",
			FileID:    uint64(newUserFile.ID),
			ObjectUrl: globalFile.ObjectName,
		})
	}

	// 合并所有分片数据
	var allData []byte
	for _, chunk := range chunks {
		allData = append(allData, chunk...)
	}

	// 生成七牛云对象名
	objectName := qiniu.GenerateObjectName(firstReq.UserID, firstReq.Filename)

	// 上传到七牛云
	qiniuClient := qiniu.NewQiniuClient()
	qiniuURL, err := qiniuClient.UploadStreamFromBytes(objectName, allData)
	if err != nil {
		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code: e.ERROR,
			Msg:  "七牛云上传失败: " + err.Error(),
		})
	}

	// 保存到数据库
	firstReq.FileSize = totalSize
	firstReq.ObjectName = qiniuURL  // 使用七牛云返回的完整URL
	file, err := dao.NewFilesDao().CreateQiniuBigFile(firstReq)
	if err != nil {
		return stream.SendAndClose(&pb.BigFileUploadResponse{
			Code: e.ERROR,
			Msg:  "数据库保存失败: " + err.Error(),
		})
	}

	return stream.SendAndClose(&pb.BigFileUploadResponse{
		Code:      e.SUCCESS,
		Msg:       "上传成功",
		FileID:    uint64(file.ID),
		ObjectUrl: qiniuURL,
	})
}

// QiniuFileDownload 七牛云文件下载（支持跨用户下载）
func (*FilesSrv) QiniuFileDownload(ctx context.Context, req *pb.FileDownloadRequest) (resp *pb.FileDownloadResponse, err error) {
	resp = new(pb.FileDownloadResponse)
	resp.Code = e.SUCCESS

	// 支持跨用户下载：如果传入了UserID，则按用户查找；否则全局查找
	var file *model.Files
	if req.UserID != 0 {
		// 按用户查找（原有逻辑）
		file, err = dao.NewFilesDao().GetFileByUIDAndFID(uint(req.UserID), uint(req.FileID))
	} else {
		// 全局查找（新增逻辑）
		file, err = dao.NewFilesDao().GetFileByID(uint(req.FileID))
	}

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

	// 检查是否为七牛云文件
	if file.Bucket != "qiniu" {
		resp.Code = e.ERROR
		resp.Msg = "该文件不是七牛云存储文件"
		return resp, nil
	}

	// 处理下载URL
	var downloadUrl string
	if strings.HasPrefix(file.FileHash, "shared_") {
		// 这是秒传文件，需要找到原始文件的URL
		// ObjectName 格式：shared_用户ID_时间戳_原始URL
		// 我们需要找到第三个下划线后的内容
		objectName := file.ObjectName

		// 找到 "shared_" 前缀后的内容
		if strings.HasPrefix(objectName, "shared_") {
			// 移除 "shared_" 前缀
			remaining := objectName[7:] // 去掉 "shared_"

			// 找到用户ID后的第一个下划线
			firstUnderscoreIndex := strings.Index(remaining, "_")
			if firstUnderscoreIndex != -1 {
				// 找到时间戳后的第二个下划线
				afterFirstUnderscore := remaining[firstUnderscoreIndex+1:]
				secondUnderscoreIndex := strings.Index(afterFirstUnderscore, "_")
				if secondUnderscoreIndex != -1 {
					// 提取原始URL（第二个下划线之后的所有内容）
					downloadUrl = afterFirstUnderscore[secondUnderscoreIndex+1:]
				} else {
					downloadUrl = file.ObjectName // 降级处理
				}
			} else {
				downloadUrl = file.ObjectName // 降级处理
			}
		} else {
			downloadUrl = file.ObjectName // 降级处理
		}
	} else {
		// 这是原始文件，直接使用 ObjectName
		downloadUrl = file.ObjectName
	}

	resp.DownloadUrl = downloadUrl
	resp.Filename = file.FileName
	resp.Msg = e.GetMsg(int(resp.Code))
	return
}

// GlobalFileSearch 全盘文件搜索
func (*FilesSrv) GlobalFileSearch(ctx context.Context, req *pb.GlobalFileSearchRequest) (resp *pb.GlobalFileSearchResponse, err error) {
	resp = new(pb.GlobalFileSearchResponse)
	resp.Code = e.SUCCESS

	// 设置默认分页参数
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大页面大小
	}

	// 调用DAO层搜索
	files, total, err := dao.NewFilesDao().GlobalFileSearch(req.FileName, page, pageSize, req.Bucket)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = "搜索文件失败: " + err.Error()
		return resp, nil
	}

	// 转换为响应格式
	var fileInfos []*pb.GlobalFileInfo
	for _, file := range files {
		// 处理下载URL（与下载接口逻辑一致）
		var downloadUrl string
		if strings.HasPrefix(file.FileHash, "shared_") {
			// 秒传文件，提取原始URL
			objectName := file.ObjectName
			if strings.HasPrefix(objectName, "shared_") {
				remaining := objectName[7:]
				firstUnderscoreIndex := strings.Index(remaining, "_")
				if firstUnderscoreIndex != -1 {
					afterFirstUnderscore := remaining[firstUnderscoreIndex+1:]
					secondUnderscoreIndex := strings.Index(afterFirstUnderscore, "_")
					if secondUnderscoreIndex != -1 {
						downloadUrl = afterFirstUnderscore[secondUnderscoreIndex+1:]
					} else {
						downloadUrl = file.ObjectName
					}
				} else {
					downloadUrl = file.ObjectName
				}
			} else {
				downloadUrl = file.ObjectName
			}
		} else {
			downloadUrl = file.ObjectName
		}

		fileInfo := &pb.GlobalFileInfo{
			FileID:     uint64(file.ID),
			FileName:   file.FileName,
			FileSize:   file.FileSize,
			Bucket:     file.Bucket,
			ObjectName: downloadUrl, // 返回可访问的URL
			FileHash:   file.FileHash,
			UserID:     uint64(file.UserID),
			CreatedAt:  file.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  file.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	resp.Files = fileInfos
	resp.Total = total
	resp.Page = page
	resp.PageSize = pageSize
	resp.Msg = e.GetMsg(int(resp.Code))
	return resp, nil
}

// QiniuFileDelete 七牛云文件删除
func (*FilesSrv) QiniuFileDelete(ctx context.Context, req *pb.FileDeleteRequest) (resp *pb.FileCommonResponse, err error) {
	resp = new(pb.FileCommonResponse)
	resp.Code = e.SUCCESS

	// 删除数据库记录
	deletedFile, err := dao.NewFilesDao().DeleteQiniuFile(uint(req.UserID), uint(req.FileID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.Code = e.ERROR
			resp.Msg = "文件不存在或无权限删除"
			return resp, nil
		}
		resp.Code = e.ERROR
		resp.Msg = "删除文件记录失败: " + err.Error()
		return resp, nil
	}

	// 检查是否需要删除七牛云上的物理文件
	shouldDeletePhysical := false

	// 如果是原始文件（不是秒传文件），需要检查是否还有其他用户在使用
	if !strings.HasPrefix(deletedFile.FileHash, "shared_") {
		// 查找相同哈希的其他文件
		sameHashFiles, err := dao.NewFilesDao().FindSameHashFiles(deletedFile.FileHash, deletedFile.ID)
		if err != nil {
			// 查询失败，为了安全起见，不删除物理文件
			log.Printf("查询相同哈希文件失败: %v", err)
		} else if len(sameHashFiles) == 0 {
			// 没有其他文件使用相同哈希，可以删除物理文件
			shouldDeletePhysical = true
		}
	}

	// 删除七牛云上的物理文件
	if shouldDeletePhysical {
		// 从ObjectName中提取七牛云的key
		objectKey := extractQiniuKey(deletedFile.ObjectName)
		if objectKey != "" {
			qiniuClient := qiniu.NewQiniuClient()
			if deleteErr := qiniuClient.DeleteFile(objectKey); deleteErr != nil {
				// 物理文件删除失败，记录日志但不影响响应
				log.Printf("删除七牛云物理文件失败: %v, key: %s", deleteErr, objectKey)
			} else {
				log.Printf("成功删除七牛云物理文件: %s", objectKey)
			}
		}
	}

	resp.Msg = e.GetMsg(int(resp.Code))
	return resp, nil
}

// extractQiniuKey 从完整URL中提取七牛云的key
func extractQiniuKey(objectName string) string {
	// 如果是完整URL，提取路径部分作为key
	if strings.HasPrefix(objectName, "http://") || strings.HasPrefix(objectName, "https://") {
		// 找到域名后的第一个斜杠
		parts := strings.SplitN(objectName, "/", 4)
		if len(parts) >= 4 {
			return parts[3] // 返回路径部分
		}
	}

	// 如果不是完整URL，直接返回
	return objectName
}
