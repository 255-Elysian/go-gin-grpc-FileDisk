package http

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"grpc-todolist-disk/app/gateway/rpc"
	"grpc-todolist-disk/app/gateway/utils"
	pb "grpc-todolist-disk/idl/pb/files"
	"grpc-todolist-disk/utils/ctl"
	"grpc-todolist-disk/utils/e"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
