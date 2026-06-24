package repository

import (
	"minlink/internal/model"

	"github.com/jinzhu/gorm"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
	List(page, limit int) ([]model.User, int, error)
	Count() (int, error)
}

// userRepository 用户数据访问实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户数据访问实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByID 根据ID查找用户
func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete 删除用户
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

// List 获取用户列表
func (r *userRepository) List(page, limit int) ([]model.User, int, error) {
	var users []model.User
	var count int

	err := r.db.Model(&model.User{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset((page - 1) * limit).Limit(limit).Order("id DESC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, count, nil
}

// Count 获取用户总数
func (r *userRepository) Count() (int, error) {
	var count int
	err := r.db.Model(&model.User{}).Count(&count).Error
	return count, err
}