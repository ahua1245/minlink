package service

import (
	"errors"
	"minlink/internal/model"
	"minlink/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// 用户服务相关错误
var (
	ErrUserNotFound       = errors.New("用户不存在")
	ErrInvalidPassword    = errors.New("密码错误")
	ErrUserDisabled       = errors.New("用户已禁用")
	ErrUsernameExists     = errors.New("用户名已存在")
	ErrInvalidOldPassword = errors.New("旧密码错误")
)

// Claims JWT 声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

// UserService 用户服务接口
type UserService interface {
	Login(username, password string) (string, *model.User, error)
	GetProfile(userID uint) (*model.User, error)
	GetProfileByUsername(username string) (*model.User, error)
	ChangePassword(userID uint, oldPassword, newPassword string) error
	UpdateProfile(userID uint, email string) error
	ListUsers(page, limit int) ([]model.User, int, error)
	GetUser(id uint) (*model.User, error)
	CreateUser(username, password, email string, role, status int) error
	UpdateUser(id uint, email string, role, status int) error
	DeleteUser(id uint) error
	InitDefaultAdmin() error
}

// userService 用户服务实现
type userService struct {
	repo      repository.UserRepository
	jwtSecret []byte
}

// NewUserService 创建用户服务实例
func NewUserService(repo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

// Login 用户登录
func (s *userService) Login(username, password string) (string, *model.User, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrUserNotFound
		}
		return "", nil, err
	}

	if user.Status != model.UserStatusActive {
		return "", nil, ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidPassword
	}

	token, err := s.generateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// generateToken 生成 JWT 令牌
func (s *userService) generateToken(userID uint, username string, role int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// GetProfile 获取用户信息
func (s *userService) GetProfile(userID uint) (*model.User, error) {
	return s.repo.FindByID(userID)
}

// GetProfileByUsername 根据用户名获取用户信息
func (s *userService) GetProfileByUsername(username string) (*model.User, error) {
	return s.repo.FindByUsername(username)
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return ErrInvalidOldPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.repo.Update(user)
}

// UpdateProfile 更新用户信息
func (s *userService) UpdateProfile(userID uint, email string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}

	user.Email = email
	return s.repo.Update(user)
}

// ListUsers 获取用户列表（管理员）
func (s *userService) ListUsers(page, limit int) ([]model.User, int, error) {
	return s.repo.List(page, limit)
}

// GetUser 获取单个用户（管理员）
func (s *userService) GetUser(id uint) (*model.User, error) {
	return s.repo.FindByID(id)
}

// UpdateUser 更新用户信息（管理员）
func (s *userService) UpdateUser(id uint, email string, role, status int) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	user.Email = email
	user.Role = role
	user.Status = status
	return s.repo.Update(user)
}

// CreateUser 创建用户（管理员）
func (s *userService) CreateUser(username, password, email string, role, status int) error {
	// 检查用户名是否已存在
	_, err := s.repo.FindByUsername(username)
	if err == nil {
		return ErrUsernameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
		Role:     role,
		Status:   status,
	}

	return s.repo.Create(user)
}

// DeleteUser 删除用户（管理员）
func (s *userService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}

// InitDefaultAdmin 初始化默认管理员
func (s *userService) InitDefaultAdmin() error {
	_, err := s.repo.FindByUsername("admin")
	if err == nil {
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.User{
		ID:       1,
		Username: "admin",
		Password: string(hashedPassword),
		Email:    "admin@example.com",
		Role:     model.UserRoleAdmin,
		Status:   model.UserStatusActive,
	}

	return s.repo.Create(admin)
}
