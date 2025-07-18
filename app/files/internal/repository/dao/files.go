package dao

import (
	"gorm.io/gorm"
	"grpc-todolist-disk/app/files/internal/repository/model"
	pb "grpc-todolist-disk/idl/pb/files"
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
