package main

import (
	"context"
	"eino-gerrit-review/internal/config"
	"eino-gerrit-review/internal/web"
	"os"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
	s := g.Server()
	port := 8000
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	s.SetPort(port)
	s.Group("/").ALL("/health", func(r *ghttp.Request) { r.Response.WriteJson(g.Map{"code": 0, "msg": "ok"}) })
	web.RegisterRoutes(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		path := os.Getenv("RULE_CONFIG_PATH")
		if path == "" {
			g.Log().Info(ctx, "RULE_CONFIG_PATH not set, skipping rule loading")
			return
		}
		g.Log().Infof(ctx, "Starting rule loader with path: %s", path)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				config.LoadRuleConfig(path)
				time.Sleep(30 * time.Second)
			}
		}
	}()

	s.Run()
}
