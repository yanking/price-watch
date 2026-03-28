package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yanking/price-watch/internal/auth"
	authconfig "github.com/yanking/price-watch/internal/auth/config"
	"github.com/yanking/price-watch/pkg/app"
	"gorm.io/gorm"
)

// HTTPServer HTTP 服务器
type HTTPServer struct {
	server *http.Server
	config Config
	logger *slog.Logger
}

// Config HTTP 服务器配置
type Config struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(
	cfg Config,
	authCfg authconfig.Config,
	db *gorm.DB,
	redisClient redis.Cmdable,
	logger *slog.Logger,
) (*HTTPServer, error) {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 引擎
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(ginLogger(logger))

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 初始化 auth 模块
	auth.InitModule(engine, db, redisClient, authCfg, logger)

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &HTTPServer{
		server: server,
		config: cfg,
		logger: logger,
	}, nil
}

// ginLogger Gin 日志中间件
func ginLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		cost := time.Since(start)
		logger.Info("HTTP request",
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", c.Writer.Status(),
			"cost", cost,
			"ip", c.ClientIP(),
			"user-agent", c.Request.UserAgent(),
			"errors", c.Errors.String(),
		)
	}
}

// Start 启动服务器
func (s *HTTPServer) Start() error {
	s.logger.Info("HTTP server starting", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *HTTPServer) Stop() error {
	s.logger.Info("HTTP server stopping")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// String 实现 app.Server 接口
func (s *HTTPServer) String() string {
	return fmt.Sprintf("HTTPServer[%s]", s.server.Addr)
}

var _ app.Server = (*HTTPServer)(nil)
