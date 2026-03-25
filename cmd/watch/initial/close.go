package initial

import "github.com/yanking/price-watch/pkg/app"

func Close(server []app.Server) (closes []app.CleanupFunc) {
	for _, s := range server {
		closes = append(closes, s.Stop)
	}

	// 其他需要关闭的资源可以在这里添加，例如数据库连接、缓存连接等

	return
}
