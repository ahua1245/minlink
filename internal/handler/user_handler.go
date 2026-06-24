package handler

import (
	"minlink/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	token, user, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "用户不存在"})
		case service.ErrInvalidPassword:
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "message": "密码错误"})
		case service.ErrUserDisabled:
			c.JSON(http.StatusForbidden, gin.H{"code": -1, "message": "用户已禁用"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "登录失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
				"status":   user.Status,
			},
		},
	})
}

// GetProfile 获取当前用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.userService.GetProfile(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取用户信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
			"status":   user.Status,
		},
	})
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	err := h.userService.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword)
	if err != nil {
		switch err {
		case service.ErrInvalidOldPassword:
			c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "旧密码错误"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "修改密码失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "密码修改成功"})
}

// UpdateProfile 更新用户信息
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	err := h.userService.UpdateProfile(userID.(uint), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

// ListUsers 获取用户列表（管理员）
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	users, total, err := h.userService.ListUsers(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取用户列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": users,
			"total": total,
		},
	})
}

// GetUser 获取单个用户（管理员）
func (h *UserHandler) GetUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	user, err := h.userService.GetUser(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    user,
	})
}

// UpdateUser 更新用户信息（管理员）
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Email  string `json:"email"`
		Role   int    `json:"role"`
		Status int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	err := h.userService.UpdateUser(uint(id), req.Email, req.Role, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

// CreateUser 创建用户（管理员）
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
		Role     int    `json:"role"`
		Status   int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	err := h.userService.CreateUser(req.Username, req.Password, req.Email, req.Role, req.Status)
	if err != nil {
		switch err {
		case service.ErrUsernameExists:
			c.JSON(http.StatusConflict, gin.H{"code": -1, "message": "用户名已存在"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "创建失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "创建成功"})
}

// DeleteUser 删除用户（管理员）
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	err := h.userService.DeleteUser(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}
