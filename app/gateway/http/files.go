package http

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"grpc-todolist-disk/app/gateway/rpc"
	"grpc-todolist-disk/app/gateway/utils"
	"grpc-todolist-disk/app/gateway/utils/mq"
	pb "grpc-todolist-disk/idl/pb/files"
	"grpc-todolist-disk/utils/ctl"
	"grpc-todolist-disk/utils/e"
	"grpc-todolist-disk/utils/qiniu"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func FileUpload(ctx *gin.Context) {
	var req pb.FileUploadRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)
	//req.ObjectName = fmt.Sprintf("%d/%d_%s", req.UserID, time.Now().UnixMilli(), utils.Clean(req.Filename))
	//log.Println("req.Filename:", req.Filename)
	//log.Println("ObjectName:", req.ObjectName)

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(200, gin.H{
			"msg":  "获取表单失败",
			"data": err.Error(),
			"code": "400",
		})
		return
	}
	files := form.File["file"]
	for _, file := range files {
		if file.Size > 10*1024*1024 { // 文件大小超过10MB
			ctx.JSON(400, gin.H{
				"msg":  "文件过大",
				"data": "文件大小超过10MB",
				"code": "400",
			})
			return
		}
		if file.Filename == "" {
			ctx.JSON(400, gin.H{
				"msg":  "上传文件名不能为空",
				"code": "400",
			})
			return
		}

		// 计算文件 hash
		src, err := file.Open()
		if err != nil {
			ctx.JSON(400, gin.H{
				"msg":  "文件打开失败",
				"data": err.Error(),
				"code": "400",
			})
			return
		}
		defer src.Close()

		h := sha256.New()
		if _, err := io.Copy(h, src); err != nil {
			ctx.JSON(400, gin.H{
				"msg":  "计算文件 Hash 失败",
				"data": err.Error(),
				"code": "400",
			})
			return
		}
		fileHash := hex.EncodeToString(h.Sum(nil))
		req.FileHash = fileHash
		// 检查数据库
		exist, err := rpc.CheckFileExists(ctx, &pb.CheckFileRequest{
			FileHash: req.FileHash,
			UserID:   req.UserID,
		})
		if err != nil {
			ctx.JSON(500, gin.H{
				"msg":  "数据库查询失败",
				"data": err.Error(),
				"code": "500",
			})
			return
		}
		if exist.Exists {
			// 命中，秒传
			ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, &pb.FileUploadResponse{
				Code:      e.SUCCESS,
				Msg:       "秒传成功，文件已存在",
				FileID:    exist.FileID,
				ObjectUrl: exist.ObjectUrl,
			}))
			return
		}

		//log.Println("file.Filename:", file.Filename)
		req.FileSize = file.Size
		//log.Println("fileSize:", req.FileSize)
		req.ObjectName = fmt.Sprintf("%d/%d_%s", req.UserID, time.Now().UnixMilli(), utils.Clean(file.Filename))
		req.Filename = file.Filename // 文件名里的中文会被”_“代替
		//log.Println("req.Filename:", req.Filename)
		filePath := filepath.Join("stores/uploaded_files", req.ObjectName)
		dir := filepath.Dir(filePath)
		//log.Println("dir:", dir)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				ctx.JSON(400, gin.H{
					"msg":  "上传失败",
					"data": err.Error(),
					"code": "400",
				})
				return
			}
		}
		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			ctx.JSON(400, gin.H{
				"msg":  "文件上传失败",
				"data": err.Error(),
				"code": "400",
			})
			return
		}
	}

	r, err := rpc.FileUpload(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "FileUpload RPC服务调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

func BigFileUpload(ctx *gin.Context) {
	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}

	// 获取文件
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "获取文件失败"))
		return
	}
	defer file.Close()

	res, err := rpc.BigFileUpload(ctx.Request.Context(), ctx.Request.Body, &rpc.UploadMeta{
		UserID:   uint64(user.ID),
		FileName: header.Filename,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "上传失败"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, res))
}

func FileList(ctx *gin.Context) {
	var req pb.FileListRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)

	r, err := rpc.FileList(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "FileList RPC服务调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

func FileDelete(ctx *gin.Context) {
	var req pb.FileDeleteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)

	r, err := rpc.FileDelete(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "FileDelete RPC服务调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

func FileDownload(ctx *gin.Context) {
	var req pb.FileDownloadRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)

	r, err := rpc.FileDownload(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "FileDownload RPC服务调用错误"))
		return
	}
	//ctx.File(r.DownloadUrl)	// 不强制下载，可以只做预览
	ctx.FileAttachment(r.DownloadUrl, r.Filename) // 强制下载
}

// AsyncFileUpload 异步上传（表单）
func AsyncFileUpload(ctx *gin.Context) {
	var req pb.FileUploadRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(200, gin.H{
			"msg":  "获取表单失败",
			"data": err.Error(),
			"code": "400",
		})
		return
	}
	files := form.File["file"]
	for _, file := range files {
		if file.Size > 10*1024*1024 { // 文件大小超过10MB
			ctx.JSON(400, gin.H{
				"msg":  "文件过大",
				"data": "文件大小超过10MB",
				"code": "400",
			})
			return
		}
		if file.Filename == "" {
			ctx.JSON(400, gin.H{
				"msg":  "上传文件名不能为空",
				"code": "400",
			})
			return
		}

		src, err := file.Open()
		if err != nil {
			ctx.JSON(200, gin.H{
				"msg":  "文件打开失败",
				"data": err.Error(),
				"code": "400",
			})
			return
		}
		defer src.Close()

		fileBytes, err := io.ReadAll(src)
		if err != nil {
			ctx.JSON(200, gin.H{
				"msg":  "读取文件失败",
				"data": err.Error(),
				"code": "400",
			})
			return
		}

		// 计算文件 hash
		hash := utils.Sha256Hash(fileBytes)
		req.FileHash = hash
		// 检查数据库
		exist, err := rpc.CheckFileExists(ctx, &pb.CheckFileRequest{
			FileHash: req.FileHash,
			UserID:   req.UserID,
		})
		if err != nil {
			ctx.JSON(500, gin.H{
				"msg":  "数据库查询失败",
				"data": err.Error(),
				"code": "500",
			})
			return
		}
		if exist.Exists {
			// 命中，秒传
			ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, &pb.FileUploadResponse{
				Code:      e.SUCCESS,
				Msg:       "秒传成功，文件已存在",
				FileID:    exist.FileID,
				ObjectUrl: exist.ObjectUrl,
			}))
			return
		}

		// 生成目标 ObjectName
		req.FileSize = file.Size
		req.ObjectName = fmt.Sprintf("%d/%d_%s", req.UserID, time.Now().UnixMilli(), utils.Clean(file.Filename))
		req.Filename = file.Filename // 文件名里的中文会被”_“代替

		// 构建 Kafka 消息体，发送到异步上传消费者
		msg := &mq.AsyncFileUploadMsg{
			UserID:     req.UserID,
			Filename:   file.Filename,
			FileSize:   file.Size,
			FileHash:   hash,
			ObjectName: req.ObjectName,
			Content:    fileBytes,
		}

		// 发送 kafka 异步任务
		if err = mq.SendFileUploadTask(msg); err != nil {
			ctx.JSON(500, gin.H{
				"msg":  "异步任务发送失败",
				"data": err.Error(),
				"code": "500",
			})
			return
		}
	}

	// 异步处理响应
	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, gin.H{
		"msg": "文件上传任务已提交",
	}))
}

// QiniuFileUpload 七牛云表单上传
func QiniuFileUpload(ctx *gin.Context) {
	var req pb.FileUploadRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(200, gin.H{
			"msg":  "获取表单失败",
			"data": err.Error(),
			"code": "400",
		})
		return
	}

	files := form.File["file"]
	for _, file := range files {
		if file.Size > 10*1024*1024 { // 文件大小超过10MB
			ctx.JSON(400, gin.H{
				"msg":  "文件过大",
				"data": "文件大小超过10MB",
				"code": "400",
			})
			return
		}
		if file.Filename == "" {
			ctx.JSON(400, gin.H{
				"msg":  "上传文件名不能为空",
				"code": "400",
			})
			return
		}

		src, err := file.Open()
		if err != nil {
			ctx.JSON(200, gin.H{
				"msg":  "文件打开失败",
				"data": err.Error(),
				"code": "400",
			})
			return
		}
		defer src.Close()

		fileBytes, err := io.ReadAll(src)
		if err != nil {
			ctx.JSON(200, gin.H{
				"msg":  "读取文件失败",
				"data": err.Error(),
				"code": "400",
			})
			return
		}

		// 计算文件 hash
		hash := utils.Sha256Hash(fileBytes)
		req.FileHash = hash
		req.FileSize = file.Size
		req.Filename = file.Filename

		// 先检查用户是否已有该文件（包括秒传记录）
		// 注意：这里不传 ObjectName，只做秒传检查
		userFileResp, err := rpc.QiniuFileUpload(ctx, &pb.FileUploadRequest{
			UserID:   req.UserID,
			Filename: req.Filename,
			FileSize: req.FileSize,
			FileHash: req.FileHash,
			// ObjectName 为空，表示只做秒传检查
		})
		if err == nil && userFileResp.Code == e.SUCCESS {
			// 用户已有该文件或全局存在相同文件，直接返回
			ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, userFileResp))
			return
		}
		// 如果返回错误且消息是"需要先上传文件到七牛云"，说明需要真正上传

		// 如果不是秒传情况，则上传到七牛云
		qiniuClient := qiniu.NewQiniuClient()
		objectName := qiniu.GenerateObjectName(req.UserID, req.Filename)
		qiniuURL, err := qiniuClient.UploadFile(objectName, fileBytes)
		if err != nil {
			ctx.JSON(500, gin.H{
				"msg":  "七牛云上传失败",
				"data": err.Error(),
				"code": "500",
			})
			return
		}

		// 保存文件信息到数据库
		req.ObjectName = qiniuURL // 存储七牛云返回的完整访问URL
		r, err := rpc.QiniuFileUpload(ctx, &req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "QiniuFileUpload RPC服务调用错误"))
			return
		}

		ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
		return
	}
}

// QiniuBigFileUpload 七牛云流式上传
func QiniuBigFileUpload(ctx *gin.Context) {
	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}

	// 获取文件
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "获取文件失败"))
		return
	}
	defer file.Close()

	res, err := rpc.QiniuBigFileUpload(ctx.Request.Context(), file, &rpc.UploadMeta{
		UserID:   uint64(user.ID),
		FileName: header.Filename,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "七牛云上传失败"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, res))
}

// QiniuFileDownload 七牛云文件下载（支持跨用户下载）
func QiniuFileDownload(ctx *gin.Context) {
	var req pb.FileDownloadRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	// 验证用户身份（需要登录才能下载）
	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}

	// 检查是否指定了 user_id 参数，如果没有则支持跨用户下载
	userIDParam := ctx.Query("user_id")
	if userIDParam != "" {
		// 如果指定了user_id，则按指定用户查找
		if userID, parseErr := strconv.ParseUint(userIDParam, 10, 64); parseErr == nil {
			req.UserID = userID
		} else {
			req.UserID = uint64(user.ID) // 解析失败则使用当前用户
		}
	} else {
		// 如果没有指定user_id，则支持跨用户下载（设置为0表示全局查找）
		req.UserID = 0
	}

	r, err := rpc.QiniuFileDownload(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "QiniuFileDownload RPC服务调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

// GlobalFileSearch 全盘文件搜索
func GlobalFileSearch(ctx *gin.Context) {
	var req pb.GlobalFileSearchRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	// 验证用户身份（虽然是全盘搜索，但仍需要登录）
	_, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}

	r, err := rpc.GlobalFileSearch(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "GlobalFileSearch RPC服务调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

// QiniuFileDelete 七牛云文件删除
func QiniuFileDelete(ctx *gin.Context) {
	var req pb.FileDeleteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	// 验证用户身份
	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)

	r, err := rpc.QiniuFileDelete(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "QiniuFileDelete RPC服务调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}
