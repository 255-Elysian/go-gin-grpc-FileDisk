package dao

import (
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"grpc-todolist-disk/app/task/internal/repository/db/model"
	pb "grpc-todolist-disk/idl/pb/task"
)

type TaskDao struct {
	*gorm.DB
}

func NewTaskDao(ctx context.Context) *TaskDao {
	return &TaskDao{NewDBClient()}
}

func (dao *TaskDao) ListTaskByUserId(userId uint) (r []*model.Task, err error) {
	err = dao.DB.Model(&model.Task{}).Where("user_id = ?", userId).Find(&r).Error
	return
}

func (dao *TaskDao) CreateTask(req *pb.TaskRequest) (err error) {
	t := &model.Task{
		UserID:    uint(req.UserID),
		Title:     req.Title,
		Content:   req.Content,
		Status:    int(req.Status),
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
	if err = dao.DB.Model(&model.Task{}).Create(&t).Error; err != nil {
		zap.Error(err)
		return
	}
	return
}

func (dao *TaskDao) DeleteTaskById(taskId, userId uint) (err error) {
	err = dao.DB.Model(&model.Task{}).Where("id = ? AND user_id = ?", taskId, userId).Delete(&model.Task{}).Error
	return
}

func (dao *TaskDao) UpdateTask(req *pb.TaskRequest) (err error) {
	taskUpdateMap := make(map[string]interface{})
	taskUpdateMap["title"] = req.Title
	taskUpdateMap["content"] = req.Content
	taskUpdateMap["status"] = int(req.Status)
	taskUpdateMap["start_time"] = req.StartTime
	taskUpdateMap["end_time"] = req.EndTime

	err = dao.DB.Model(&model.Task{}).Where("id = ?", req.TaskID).Updates(&taskUpdateMap).Error
	return
}

func (dao *TaskDao) ShowTaskById(taskId, userId uint) (r []*model.Task, err error) {
	err = dao.DB.Model(&model.Task{}).Where("id = ? AND user_id = ?", taskId, userId).Find(&r).Error
	return
}
