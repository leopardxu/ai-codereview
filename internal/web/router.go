package web

import (
    "github.com/gogf/gf/v2/net/ghttp"
)

func RegisterRoutes(s *ghttp.Server) {
    group := s.Group("/")
    group.GET("/changes", GetChanges)
    group.POST("/reviews/run", RunReview)
    group.GET("/reviews/{id}", GetReview)
    group.POST("/reviews/{id}/publish", PublishReview)
    group.POST("/scheduler/scan", TriggerScan)
    group.GET("/metrics", Metrics)
    group.POST("/config/rules/reload", ReloadRules)
}
