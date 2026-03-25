package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GinServer struct {
	engine *gin.Engine
	server *http.Server
}

func NewGinServer(addr string, setupRoutes func(*gin.Engine)) *GinServer {
	engine := gin.New()
	engine.Use(gin.Recovery())
	setupRoutes(engine)

	return &GinServer{
		engine: engine,
		server: &http.Server{
			Addr:    addr,
			Handler: engine,
		},
	}
}

func (s *GinServer) Start() error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

func (s *GinServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *GinServer) String() string { return "http-server" }
