package initial

import (
	"time"

	"github.com/yanking/price-watch/internal/watch/server"
	"github.com/yanking/price-watch/internal/watch/svc"
	"github.com/yanking/price-watch/pkg/app"
)

// CreateServices 创建应用所需的所有服务器
// 参数 ctx 为服务上下文，包含配置、日志器等依赖
// 返回 servers 列表，可包含 HTTP 服务器、GRPC 服务器等
func CreateServices(ctx *svc.ServiceContext) (services []app.Server) {
	// 创建 HTTP 服务器配置
	httpCfg := server.Config{
		Host:         "0.0.0.0",
		Port:         8080,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 创建 HTTP 服务器
	httpServer, err := server.NewHTTPServer(httpCfg, ctx.AuthConfig, ctx.MySQL, ctx.Logger)
	if err != nil {
		ctx.Logger.Error("create http server", "error", err)
		return
	}

	services = append(services, httpServer)
	return
}
