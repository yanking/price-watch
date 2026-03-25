package initial

import "github.com/yanking/price-watch/pkg/app"

func Close(servers []app.Server) (closes []app.CleanupFunc) {
	// Stop servers in reverse order: adapters first, then HTTP
	for i := len(servers) - 1; i >= 0; i-- {
		s := servers[i]
		closes = append(closes, s.Stop)
	}
	return
}
