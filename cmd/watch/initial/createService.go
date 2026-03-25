package initial

import (
	"github.com/yanking/price-watch/internal/watch/svc"
	"github.com/yanking/price-watch/pkg/app"
)

// CreateServices 创建应用所需的所有服务器
// 参数 ctx 为服务上下文，包含配置、日志器等依赖
// 返回 servers 列表，可包含 HTTP 服务器、GRPC 服务器等
func CreateServices(ctx *svc.ServiceContext) (services []app.Server) {
	// TODO: 创建服务器，例如：
	// services = append(services, NewHTTPServer(ctx.Config.HTTP, ctx.Logger))
	// services = append(services, NewGRPCServer(ctx.Config.GRPC, ctx.Logger))
	return
}
