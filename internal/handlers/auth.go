package handlers

import (
	"net/http"

	"github.com/Pokem0n2/gym-log/internal/middleware"
	"github.com/Pokem0n2/gym-log/internal/models"
	"github.com/Pokem0n2/gym-log/internal/repository"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	store *repository.UserStore
}

func NewAuthHandler(store *repository.UserStore) *AuthHandler {
	return &AuthHandler{store: store}
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.store.ValidateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := middleware.GenerateToken(user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token 生成失败"})
		return
	}

	// HttpOnly Cookie，前端无需处理 token
	c.SetCookie("token", token, 7*24*3600, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"user_id":  user.UserID,
		"username": user.Username,
	})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "已登出"})
}

// Me 获取当前登录用户信息
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.store.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}
