package handler

import (
	"net/http"
	"strconv"

	"minlink/internal/service"

	"github.com/gin-gonic/gin"
)

type ShortURLHandler struct {
	service service.ShortURLService
}

func NewShortURLHandler(service service.ShortURLService) *ShortURLHandler {
	return &ShortURLHandler{service: service}
}

func (h *ShortURLHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("shortCode")

	longURL, err := h.service.GetLongURL(shortCode)
	if err != nil {
		if err == service.ErrShortCodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "短链不存在"})
			return
		}
		if err == service.ErrShortCodeExpired {
			c.JSON(http.StatusGone, gin.H{"code": 410, "message": "短链已过期"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务器内部错误"})
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	go h.service.RecordVisit(shortCode, ip, userAgent, referer)

	c.Redirect(http.StatusFound, longURL)
}

func (h *ShortURLHandler) CreateShortURL(c *gin.Context) {
	var req struct {
		LongURL    string `json:"long_url" binding:"required"`
		ExpireDays int    `json:"expire_days"`
		Name       string `json:"name"`
		Remark     string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 获取当前用户信息（游客为 nil）
	var userID uint = 0
	var username string = ""

	userIDValue, exists := c.Get("user_id")
	if exists {
		userID = userIDValue.(uint)
	}

	usernameValue, exists := c.Get("username")
	if exists {
		username = usernameValue.(string)
	}

	// 游客权限校验：只能创建一周内的短链
	if userID == 0 && req.ExpireDays > 7 && req.ExpireDays != 0 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "游客只能创建有效期一周内的短链"})
		return
	}

	shortURL, err := h.service.CreateShortURL(req.LongURL, req.ExpireDays, userID, username, req.Name, req.Remark)
	if err != nil {
		if err == service.ErrInvalidURL {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的URL"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	host := c.Request.Host
	shortURLStr := "http://" + host + "/" + shortURL.ShortCode

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"short_url":  shortURLStr,
			"short_code": shortURL.ShortCode,
			"expire_at":  shortURL.ExpireAt,
			"created_by": shortURL.CreatedBy,
		},
	})
}

func (h *ShortURLHandler) GetStats(c *gin.Context) {
	shortCode := c.Param("code")

	stats, err := h.service.GetStats(shortCode)
	if err != nil {
		if err == service.ErrShortCodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "短链不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": stats,
	})
}

func (h *ShortURLHandler) ListShortURLs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	shortURLs, total, err := h.service.ListShortURLs(0, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"items": shortURLs,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *ShortURLHandler) UpdateStatus(c *gin.Context) {
	shortCode := c.Param("code")

	var req struct {
		Status int `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	err := h.service.UpdateStatus(shortCode, req.Status)
	if err != nil {
		if err == service.ErrShortCodeNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "短链不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

func (h *ShortURLHandler) DeleteShortURL(c *gin.Context) {
	shortCode := c.Param("code")

	err := h.service.DeleteShortURL(shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ListUserShortURLs 用户查看自己的短链列表
func (h *ShortURLHandler) ListUserShortURLs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	shortURLs, total, err := h.service.ListShortURLs(userID.(uint), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"items": shortURLs,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// ListAdminShortURLs 管理员查看所有短链列表
func (h *ShortURLHandler) ListAdminShortURLs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	// userID = 0 表示查询所有
	shortURLs, total, err := h.service.ListShortURLs(0, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"items": shortURLs,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}
