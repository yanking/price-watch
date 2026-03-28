package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 常用响应码
const (
	CodeSuccess      = 0
	CodeError        = 1
	CodeInvalidParam = 40001
	CodeUnauthorized = 40101
	CodeForbidden    = 40301
	CodeNotFound     = 40401
	CodeServerError  = 50001
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	var httpStatus int
	switch {
	case code >= CodeServerError:
		httpStatus = http.StatusInternalServerError
	case code >= CodeNotFound:
		httpStatus = http.StatusNotFound
	case code >= CodeForbidden:
		httpStatus = http.StatusForbidden
	case code >= CodeUnauthorized:
		httpStatus = http.StatusUnauthorized
	case code >= CodeInvalidParam:
		httpStatus = http.StatusBadRequest
	default:
		httpStatus = http.StatusOK
	}

	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithMessage 自定义错误消息
func ErrorWithMessage(c *gin.Context, message string) {
	Error(c, CodeError, message)
}

// BadRequest 请求参数错误
func BadRequest(c *gin.Context, message string) {
	Error(c, CodeInvalidParam, message)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, message string) {
	Error(c, CodeUnauthorized, message)
}

// Forbidden 无权限
func Forbidden(c *gin.Context, message string) {
	Error(c, CodeForbidden, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	Error(c, CodeNotFound, message)
}

// ServerError 服务器错误
func ServerError(c *gin.Context, message string) {
	Error(c, CodeServerError, message)
}
