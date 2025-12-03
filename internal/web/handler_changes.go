package web

import (
    "context"
    "eino-gerrit-review/internal/app/eino/core"
    "eino-gerrit-review/internal/app/eino/flows"
    "eino-gerrit-review/internal/app/scheduler"
    "eino-gerrit-review/internal/app/tools"

    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/net/ghttp"
)

func GetChanges(r *ghttp.Request) {
	f := &flows.IngestFlow{}
	res, _ := f.Execute(context.Background(), core.NewFlowContext())
	r.Response.WriteJson(g.Map{"code": 0, "data": res["changes"]})
}

func TriggerScan(r *ghttp.Request) {
    project := r.Get("project").String()
    branch := r.Get("branch").String()
    if project == "" || branch == "" {
        r.Response.WriteJson(g.Map{"code":1,"msg":"missing project/branch"}); return
    }
    if !validParam(project) || !validParam(branch) { r.Response.WriteJson(g.Map{"code":1,"msg":"invalid project/branch"}); return }
    gt := &tools.GerritTool{}
    changes, _ := gt.GetOpenChanges(project, branch, 10)
    pool := scheduler.NewWorkerPool(8)
    pool.Run(context.Background())
    for _, c := range changes {
        id := c["id"].(string)
        pool.Submit(scheduler.Task{ChangeId: id, Patchset: "1", EnableContext: r.Get("enableContext").Bool()})
    }
    r.Response.WriteJson(g.Map{"code": 0, "data": g.Map{"scanned": len(changes), "queued": len(changes)}})
}
