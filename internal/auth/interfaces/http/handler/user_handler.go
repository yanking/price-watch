package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/auth/application/dto"
	"github.com/yanking/price-watch/internal/auth/application/service"
	"github.com/yanking/price-watch/internal/auth/interfaces/http/middleware"
	"github.com/yanking/price-watch/internal/auth/interfaces/http/response"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Router /api/v1/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	resp, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新当前登录用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.UpdateProfileRequest true "更新信息"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Router /api/v1/user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Response
// @Router /api/v1/user/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.Success(c, nil)
}
