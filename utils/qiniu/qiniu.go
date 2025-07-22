package qiniu

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"grpc-todolist-disk/conf"
)

type QiniuClient struct {
	mac        *qbox.Mac                 // 鉴权用的 MAC 实例（AK/SK）
	cfg        *storage.Config           // 存储配置，包括区域、是否使用 HTTPS
	bucket     string                    // 存储空间名
	domain     string                    // 对外访问域名（用于拼接 URL）
	uploader   *storage.FormUploader     // 表单上传工具
	resumeUpV2 *storage.ResumeUploaderV2 // 分片上传工具（适用于流式上传）
}

// NewQiniuClient 创建七牛云客户端
func NewQiniuClient() *QiniuClient {
	qiniuConf := conf.Conf.Qiniu
	mac := qbox.NewMac(qiniuConf.AccessKey, qiniuConf.SecretKey)

	cfg := &storage.Config{
		UseHTTPS:      false,
		UseCdnDomains: false,
	}

	// 根据配置设置存储区域
	switch qiniuConf.Zone {
	case "z0":
		cfg.Zone = &storage.ZoneHuadong
	case "z1":
		cfg.Zone = &storage.ZoneHuabei
	case "z2":
		cfg.Zone = &storage.ZoneHuanan
	case "na0":
		cfg.Zone = &storage.ZoneBeimei
	case "as0":
		cfg.Zone = &storage.ZoneXinjiapo
	default:
		cfg.Zone = &storage.ZoneHuadong // 默认华东
	}

	return &QiniuClient{
		mac:        mac,
		cfg:        cfg,
		bucket:     qiniuConf.Bucket,
		domain:     qiniuConf.Domain,
		uploader:   storage.NewFormUploader(cfg),
		resumeUpV2: storage.NewResumeUploaderV2(cfg),
	}
}

// getUploadToken 获取上传凭证
func (q *QiniuClient) getUploadToken(key string) string {
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", q.bucket, key),
	}
	return putPolicy.UploadToken(q.mac)
}

// UploadFile 表单上传文件
func (q *QiniuClient) UploadFile(key string, data []byte) (string, error) {
	upToken := q.getUploadToken(key)
	ret := storage.PutRet{}

	err := q.uploader.Put(context.Background(), &ret, upToken, key, bytes.NewReader(data), int64(len(data)), nil)
	if err != nil {
		return "", fmt.Errorf("七牛云上传失败: %w", err)
	}

	// 返回文件访问URL
	return q.getFileURL(ret.Key), nil
}

// UploadStream 流式上传文件
func (q *QiniuClient) UploadStream(key string, reader io.Reader, size int64) (string, error) {
	upToken := q.getUploadToken(key)
	ret := storage.PutRet{}

	// 将 io.Reader 转换为 io.ReaderAt
	// 先读取所有数据到内存中
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取数据失败: %w", err)
	}

	// 使用 bytes.Reader 实现 io.ReaderAt 接口
	readerAt := bytes.NewReader(data)

	// 使用分片上传
	err = q.resumeUpV2.Put(context.Background(), &ret, upToken, key, readerAt, int64(len(data)), nil)
	if err != nil {
		return "", fmt.Errorf("七牛云流式上传失败: %w", err)
	}

	// 返回文件访问URL
	return q.getFileURL(ret.Key), nil
}

// UploadStreamFromBytes 从字节数据流式上传文件（适用于已在内存中的数据）
func (q *QiniuClient) UploadStreamFromBytes(key string, data []byte) (string, error) {
	upToken := q.getUploadToken(key)
	ret := storage.PutRet{}

	// 使用 bytes.Reader 实现 io.ReaderAt 接口
	readerAt := bytes.NewReader(data)

	// 使用分片上传
	err := q.resumeUpV2.Put(context.Background(), &ret, upToken, key, readerAt, int64(len(data)), nil)
	if err != nil {
		return "", fmt.Errorf("七牛云流式上传失败: %w", err)
	}

	// 返回文件访问URL
	return q.getFileURL(ret.Key), nil
}

// getFileURL 获取文件访问URL
func (q *QiniuClient) getFileURL(key string) string {
	// 如果key已经是完整的URL，直接返回
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}

	// 如果没有配置域名，返回key
	if q.domain == "" {
		return key
	}

	// 构造完整的URL
	return fmt.Sprintf("http://%s/%s", q.domain, key)
}

// GenerateObjectName 生成对象名称
func GenerateObjectName(userID uint64, filename string) string {
	ext := filepath.Ext(filename)
	timestamp := time.Now().UnixMilli()
	return fmt.Sprintf("uploads/%d/%d%s", userID, timestamp, ext)
}

// DeleteFile 删除文件
func (q *QiniuClient) DeleteFile(key string) error {
	bucketManager := storage.NewBucketManager(q.mac, q.cfg)
	err := bucketManager.Delete(q.bucket, key)
	if err != nil {
		return fmt.Errorf("删除七牛云文件失败: %w", err)
	}
	return nil
}
