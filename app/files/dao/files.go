package dao

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"grpc-todolist-disk/app/files/internal/repository/model"
	pb "grpc-todolist-disk/idl/pb/files"
	"strings"
	"time"
)

type FilesDao struct {
	*gorm.DB
}

func NewFilesDao() *FilesDao {
	return &FilesDao{
		NewDBClient(),
	}
}

func (dao *FilesDao) CreateFile(req *pb.FileUploadRequest) (*model.Files, error) {
	file := &model.Files{
		Model:      gorm.Model{},
		UserID:     uint(req.UserID),
		FileName:   req.Filename,
		FileSize:   req.FileSize,
		Bucket:     "local",
		ObjectName: req.ObjectName,
		FileHash:   req.FileHash,
	}
	if err := dao.DB.Model(&model.Files{}).Create(&file).Error; err != nil {
		return nil, err
	}
	return file, nil
}

func (dao *FilesDao) CreateBigFile(req *pb.BigFileUploadRequest) (*model.Files, error) {
	file := &model.Files{
		Model:      gorm.Model{},
		UserID:     uint(req.UserID),
		FileName:   req.Filename,
		FileSize:   req.FileSize,
		Bucket:     "local",
		ObjectName: req.ObjectName,
		FileHash:   req.FileHash,
	}
	if err := dao.DB.Model(&model.Files{}).Create(&file).Error; err != nil {
		return nil, err
	}
	return file, nil
}

func (dao *FilesDao) ListFiles(req *pb.FileListRequest) (f []*model.Files, total int64, err error) {
	query := dao.DB.Model(&model.Files{}).Where("user_id = ?", req.UserID)
	err = query.Count(&total).Error
	if err != nil {
		return
	}
	err = query.Offset(int((req.Page - 1) * req.PageSize)).Limit(int(req.PageSize)).Find(&f).Error
	return
}

func (dao *FilesDao) DeleteFile(req *pb.FileDeleteRequest) error {
	return dao.DB.Model(&model.Files{}).Where("id = ? AND user_id = ?", req.FileID, req.UserID).Delete(&model.Files{}).Error
}

func (dao *FilesDao) GetFileByUIDAndFID(uID, fID uint) (f *model.Files, err error) {
	err = dao.DB.Model(&model.Files{}).Where("id = ? AND user_id = ?", fID, uID).First(&f).Error
	return
}

// FindByHash 秒传哈希检测 - 检查当前用户是否已有该文件
func (dao *FilesDao) FindByHash(req *pb.CheckFileRequest) (*model.Files, error) {
	var file model.Files
	// 检查当前用户是否已经有这个文件的记录
	// 排除以"shared_"开头的伪造哈希值，只匹配真实的文件哈希
	err := dao.DB.Model(&model.Files{}).Where("user_id = ? AND file_hash = ? AND file_hash NOT LIKE 'shared_%'", req.UserID, req.FileHash).First(&file).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &file, err
}

// FindUserFileByOriginalHash 检查用户是否已有基于某个原始哈希的文件记录（包括秒传记录）
func (dao *FilesDao) FindUserFileByOriginalHash(userID uint64, originalHash string) (*model.Files, error) {
	var file model.Files

	// 首先查找用户是否有真实哈希的记录
	err := dao.DB.Model(&model.Files{}).Where("user_id = ? AND file_hash = ?", userID, originalHash).First(&file).Error
	if err == nil {
		return &file, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 然后查找是否有基于该原始文件的秒传记录
	// 先查找全局原始文件
	var originalFile model.Files
	err = dao.DB.Model(&model.Files{}).Where("file_hash = ? AND file_hash NOT LIKE 'shared_%'", originalHash).First(&originalFile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // 全局都没有这个文件
	}
	if err != nil {
		return nil, err
	}

	// 查找用户是否有指向同一个原始 ObjectName 的秒传记录
	// 秒传记录的 ObjectName 格式为: shared_用户ID_时间戳_原始ObjectName
	var userSharedFile model.Files
	err = dao.DB.Model(&model.Files{}).Where("user_id = ? AND object_name LIKE ? AND file_hash LIKE 'shared_%'",
		userID, "%"+originalFile.ObjectName).First(&userSharedFile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &userSharedFile, err
}

// FindGlobalByHash 全局检查是否存在相同哈希的文件（用于跨用户秒传）
func (dao *FilesDao) FindGlobalByHash(fileHash string) (*model.Files, error) {
	var file model.Files
	// 只查找真实的文件哈希，排除以"shared_"开头的伪造哈希值
	err := dao.DB.Model(&model.Files{}).Where("file_hash = ? AND file_hash NOT LIKE 'shared_%'", fileHash).First(&file).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &file, err
}

// CreateUserFileFromExisting 为用户创建基于已存在文件的新记录
func (dao *FilesDao) CreateUserFileFromExisting(userID uint64, filename string, existingFile *model.Files) (*model.Files, error) {
	// 为了避免 ObjectName 重复，我们在原有基础上添加用户ID和时间戳
	// 但在实际应用中，我们知道这指向的是同一个物理文件
	timestamp := time.Now().UnixMilli()
	uniqueObjectName := fmt.Sprintf("shared_%d_%d_%s", userID, timestamp, existingFile.ObjectName)

	// 为了避免 FileHash 重复（空字符串冲突），生成一个唯一的标识
	// 格式：shared_用户ID_时间戳，这样确保每个用户的秒传记录都有唯一的hash标识
	uniqueFileHash := fmt.Sprintf("shared_%d_%d", userID, timestamp)

	userFile := &model.Files{
		UserID:     uint(userID),
		FileName:   filename,
		FileSize:   existingFile.FileSize,
		Bucket:     existingFile.Bucket,
		ObjectName: uniqueObjectName, // 使用唯一的对象名
		FileHash:   uniqueFileHash,   // 使用唯一的哈希标识
	}

	if err := dao.DB.Create(&userFile).Error; err != nil {
		return nil, err
	}
	return userFile, nil
}

// CreateQiniuFile 创建七牛云文件记录
func (dao *FilesDao) CreateQiniuFile(req *pb.FileUploadRequest) (*model.Files, error) {
	file := &model.Files{
		Model:      gorm.Model{},
		UserID:     uint(req.UserID),
		FileName:   req.Filename,
		FileSize:   req.FileSize,
		Bucket:     "qiniu",
		ObjectName: req.ObjectName,
		FileHash:   req.FileHash,
	}

	if err := dao.DB.Model(&model.Files{}).Create(&file).Error; err != nil {
		return nil, err
	}
	return file, nil
}

// CreateQiniuBigFile 创建七牛云大文件记录
func (dao *FilesDao) CreateQiniuBigFile(req *pb.BigFileUploadRequest) (*model.Files, error) {
	file := &model.Files{
		Model:      gorm.Model{},
		UserID:     uint(req.UserID),
		FileName:   req.Filename,
		FileSize:   req.FileSize,
		Bucket:     "qiniu",
		ObjectName: req.ObjectName,
		FileHash:   req.FileHash,
	}

	if err := dao.DB.Model(&model.Files{}).Create(&file).Error; err != nil {
		return nil, err
	}
	return file, nil
}

// GlobalFileSearch 全盘文件搜索
func (dao *FilesDao) GlobalFileSearch(fileName string, page, pageSize uint32, bucket string) ([]*model.Files, uint32, error) {
	var files []*model.Files
	var total int64

	// 构建查询条件
	query := dao.DB.Model(&model.Files{})

	// 文件名模糊搜索
	if fileName != "" {
		query = query.Where("file_name LIKE ?", "%"+fileName+"%")
	}

	// 存储桶过滤
	if bucket != "" {
		query = query.Where("bucket = ?", bucket)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, uint32(total), nil
}

// GetFileByID 根据文件ID获取文件信息（不限制用户）
func (dao *FilesDao) GetFileByID(fileID uint) (*model.Files, error) {
	var file model.Files
	err := dao.DB.Model(&model.Files{}).Where("id = ?", fileID).First(&file).Error
	return &file, err
}

// DeleteQiniuFile 删除七牛云文件记录
func (dao *FilesDao) DeleteQiniuFile(userID, fileID uint) (*model.Files, error) {
	var file model.Files

	// 先查找文件
	err := dao.DB.Model(&model.Files{}).Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error
	if err != nil {
		return nil, err
	}

	// 检查是否为七牛云文件
	if file.Bucket != "qiniu" {
		return nil, fmt.Errorf("该文件不是七牛云存储文件")
	}

	// 删除数据库记录
	err = dao.DB.Delete(&file).Error
	if err != nil {
		return nil, err
	}

	return &file, nil
}

// FindSameHashFiles 查找相同哈希的其他文件（用于判断是否需要删除物理文件）
func (dao *FilesDao) FindSameHashFiles(fileHash string, excludeFileID uint) ([]*model.Files, error) {
	var files []*model.Files

	// 如果是秒传文件（FileHash以shared_开头），不需要删除物理文件
	if strings.HasPrefix(fileHash, "shared_") {
		return files, nil
	}

	// 查找相同真实哈希的其他文件
	err := dao.DB.Model(&model.Files{}).Where("file_hash = ? AND id != ? AND file_hash NOT LIKE 'shared_%'",
		fileHash, excludeFileID).Find(&files).Error

	return files, err
}
