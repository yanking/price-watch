package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/auth/application/dto"
	"github.com/yanking/price-watch/internal/auth/application/service"
	"github.com/yanking/price-watch/internal/auth/interfaces/http/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register 注册
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册信息"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// Login 登录
// @Summary 用户登录
// @Description 使用用户名/邮箱/手机号登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// OAuthLogin OAuth 登录
// @Summary OAuth 登录
// @Description 使用第三方账号登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.OAuthLoginRequest true "OAuth登录信息"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /api/v1/auth/oauth [post]
func (h *AuthHandler) OAuthLogin(c *gin.Context) {
	var req dto.OAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.authService.OAuthLogin(c.Request.Context(), &req)
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.Success(c, resp)
}
