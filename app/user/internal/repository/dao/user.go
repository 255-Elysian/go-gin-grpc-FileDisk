package dao

import (
	"gorm.io/gorm"
	"grpc-todolist-disk/app/user/internal/repository/cache"
	"grpc-todolist-disk/app/user/internal/repository/model"
	pb "grpc-todolist-disk/idl/pb/user"
	"log"
	"time"
)

type UserDao struct {
	*gorm.DB
}

func NewUserDao() *UserDao {
	return &UserDao{
		NewDBClient(),
	}
}

func GetUserByName(name string) *model.User {
	var user *model.User
	user = cache.GetUserFromRedisByName(name)
	// 如果Redis中没有缓存，则查询MySQL数据库
	if user != nil {
		return user
	}

	lock := cache.GetRedLock(name)
	if err := lock.Lock(); err != nil {
		log.Println("获取锁失败:", err)
		return nil
	}
	// 启动续租协程，并可控制退出
	stop := make(chan struct{})
	go cache.ContinueLock(lock, stop)
	defer func() {
		close(stop)          // 通知协程退出
		_, _ = lock.Unlock() // 主动解锁
	}()

	db := NewDBClient().Session(&gorm.Session{NewDB: true})
	user = &model.User{}
	err := db.Where("username = ?", name).First(user).Error
	//log.Printf("数据库查询结果：%+v\n", user)
	if err != nil {
		log.Println(err)
		return nil
	}
	cache.SetNameToRedis(name, user)
	return user
}

func SetUserStatus(user *model.User, status string) bool {
	// 方式一：分布式锁，强一致性
	lock := cache.GetRedLock(user.Username)
	// 尝试获取锁
	if err := lock.Lock(); err != nil {
		log.Println("获取锁失败:", err)
		return false
	}
	// 启动续租协程，并添加退出控制
	stop := make(chan struct{})
	go cache.ContinueLock(lock, stop)

	// 函数退出时释放锁 + 通知续租停止
	defer func() {
		close(stop)
		_, _ = lock.Unlock()
	}()

	db := NewDBClient().Session(&gorm.Session{NewDB: true})
	user.Status = status
	cache.ClearNameRedisCache(user.Username)
	err := db.Save(&user).Error
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func SetUserPassword(user *model.User, newPassword string) bool {
	db := NewDBClient().Session(&gorm.Session{NewDB: true})
	_ = user.SetPassword(newPassword)
	// 方式二：双删，弱一致性
	go cache.ClearNameRedisCache(user.Username)
	err := db.Save(&user).Error
	if err != nil {
		log.Println(err)
		return false
	}
	go cache.SendMsgToMq(user.Username, time.Now().Add(1000*time.Millisecond))
	return true
}

func RecordPasswordWrong(user *model.User, tires uint) bool {
	// 方式一：分布式锁，强一致性
	lock := cache.GetRedLock(user.Username)
	// 尝试获取锁
	if err := lock.Lock(); err != nil {
		log.Println(err)
		return false
	}
	// 启动续租协程，并添加退出控制
	stop := make(chan struct{})
	go cache.ContinueLock(lock, stop)

	// 函数退出时释放锁 + 通知续租停止
	defer func() {
		close(stop)
		_, _ = lock.Unlock()
	}()

	db := NewDBClient().Session(&gorm.Session{NewDB: true})
	user.PasswordTry = tires
	if user.PasswordTry >= 10 {
		user.PasswordTry = 0
		user.LockedUntil = time.Now().Add(time.Hour)
	}
	cache.ClearNameRedisCache(user.Username)
	err := db.Save(&user).Error
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (dao *UserDao) CreateUser(req *pb.UserRequest) error {
	user := &model.User{
		Model:       gorm.Model{},
		Username:    req.Username,
		Nickname:    req.Nickname,
		Password:    req.Password,
		PasswordTry: 0,
		LockedUntil: time.Now(),
	}
	_ = user.SetPassword(user.Password)
	return dao.DB.Model(&model.User{}).Create(&user).Error
}

func (dao *UserDao) DeleteUser(name string) error {
	cache.ClearNameRedisCache(name)
	return dao.DB.Where("username = ?", name).Delete(&model.User{}).Error
}

func (dao *UserDao) GetUserByUserID(uID uint) (user *model.User, err error) {
	err = dao.DB.Model(&model.User{}).Where("id = ?", uID).First(&user).Error
	return
}
