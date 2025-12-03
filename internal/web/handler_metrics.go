package web

import (
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/net/ghttp"
    "eino-gerrit-review/internal/monitor"
)

func Metrics(r *ghttp.Request) {
    r.Response.WriteJson(g.Map{"code": 0, "data": g.Map{
        "node_calls": monitor.NodeCalls,
        "node_errors": monitor.NodeErrors,
        "context_calls": monitor.ContextCalls,
        "context_cache_hits": monitor.ContextCacheHits,
        "context_cache_miss": monitor.ContextCacheMiss,
        "graph_exec_millis_sum": monitor.GraphExecMillisSum,
        "graph_exec_count": monitor.GraphExecCount,
        "context_func_count": monitor.ContextFuncCount,
        "context_class_count": monitor.ContextClassCount,
        "context_dep_count": monitor.ContextDepCount,
    }})
}
