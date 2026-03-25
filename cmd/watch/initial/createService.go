package initial

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/watch/exchange"
	"github.com/yanking/price-watch/internal/watch/handler"
	"github.com/yanking/price-watch/internal/watch/server"
	"github.com/yanking/price-watch/internal/watch/svc"
	"github.com/yanking/price-watch/pkg/app"
)

func CreateServices(ctx *svc.ServiceContext) (services []app.Server) {
	httpSrv := server.NewGinServer(ctx.Config.HTTP.Addr, func(e *gin.Engine) {
		handler.RegisterRoutes(e, ctx.Redis.Client(), ctx.Influx, ctx.SubMgr, ctx.Adapters)
	})
	services = append(services, httpSrv)

	for _, adapter := range ctx.Adapters {
		services = append(services, &adapterServer{adapter: adapter})
	}

	return
}

type adapterServer struct {
	adapter exchange.ExchangeAdapter
}

func (s *adapterServer) Start() error   { return s.adapter.Start(context.Background()) }
func (s *adapterServer) Stop() error    { return s.adapter.Stop() }
func (s *adapterServer) String() string { return fmt.Sprintf("exchange-%s", s.adapter.Name()) }
