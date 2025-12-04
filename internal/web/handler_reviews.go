package web

import (
	"context"
	"eino-gerrit-review/internal/app/eino"
	"eino-gerrit-review/internal/app/eino/core"
	"eino-gerrit-review/internal/app/eino/flows"
	"eino-gerrit-review/internal/app/tools"

	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type RunReviewReq struct {
	ChangeNum     string `json:"changeNum"`
	Patchset      string `json:"patchset"`
	EnableContext bool   `json:"enableContext"`
	React         bool   `json:"react"`
	AutoPublish   bool   `json:"autoPublish"`
}

func RunReview(r *ghttp.Request) {
	var req RunReviewReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": 1, "msg": "invalid request"})
		return
	}
	if req.ChangeNum == "" || req.Patchset == "" {
		r.Response.WriteJson(g.Map{"code": 1, "msg": "missing changeNum or patchset"})
		return
	}
	// 简易校验：只允许字母数字、-、_，长度限制
	if len(req.ChangeNum) > 64 || len(req.Patchset) > 32 {
		r.Response.WriteJson(g.Map{"code": 1, "msg": "param too long"})
		return
	}
	if !validParam(req.ChangeNum) || !validParam(req.Patchset) {
		r.Response.WriteJson(g.Map{"code": 1, "msg": "invalid param format"})
		return
	}
	f := &flows.ReviewFlow{}
	fc := core.NewFlowContext()
	fc.ChangeNum = req.ChangeNum
	fc.Patchset = req.Patchset
	fc.EnableContext = req.EnableContext
	var res core.Result
	if req.React {
		// 使用 React 编排
		rgraph, err := eino.BuildReactGraph()
		if err != nil {
			r.Response.WriteJson(g.Map{"code": 1, "msg": "build react graph failed: " + err.Error()})
			return
		}
		runnable, err := rgraph.Compile(context.Background())
		if err != nil {
			r.Response.WriteJson(g.Map{"code": 1, "msg": "compile react graph failed"})
			return
		}
		out, err := runnable.Invoke(context.Background(), map[string]any{"changeNum": req.ChangeNum, "patchset": req.Patchset, "enableContext": req.EnableContext})
		if err != nil {
			r.Response.WriteJson(g.Map{"code": 1, "msg": "invoke react graph failed"})
			return
		}
		res = core.Result{"preview": out["preview"]}
	} else {
		var err error
		res, err = f.Execute(context.Background(), fc)
		if err != nil {
			r.Response.WriteJson(g.Map{"code": 1, "msg": "execute review flow failed: " + err.Error()})
			return
		}
	}
	id := "R" + toStr(time.Now().UnixNano())
	if v, ok := res["preview"].(map[string]interface{}); ok {
		core.PutReview(id, v, req.ChangeNum, req.Patchset)

		// Check for AutoPublish
		if req.AutoPublish {
			gt := &tools.GerritTool{}
			if _, err := gt.PostReview(req.ChangeNum, req.Patchset, v); err != nil {
				// If publish fails, we still return the reviewId but with a warning or error msg
				// For now, let's just log it or include in response
				r.Response.WriteJson(g.Map{"code": 0, "data": g.Map{"reviewId": id, "preview": res["preview"], "published": false, "publishError": err.Error()}})
				return
			}
			r.Response.WriteJson(g.Map{"code": 0, "data": g.Map{"reviewId": id, "preview": res["preview"], "published": true}})
			return
		}
	}
	r.Response.WriteJson(g.Map{"code": 0, "data": g.Map{"reviewId": id, "preview": res["preview"]}})
}

func toStr(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [25]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func validParam(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '-' || b == '_' || b == '.' {
			continue
		}
		return false
	}
	return true
}

func GetReview(r *ghttp.Request) {
	id := r.Get("id").String()
	if v, ok := core.GetReview(id); ok {
		r.Response.WriteJson(g.Map{"code": 0, "data": g.Map{"reviewId": id, "preview": v.Payload, "changeNum": v.ChangeNum, "patchset": v.Patchset}})
		return
	}
	r.Response.WriteJson(g.Map{"code": 1, "msg": "not found"})
}

func PublishReview(r *ghttp.Request) {
	id := r.Get("id").String()
	v, ok := core.GetReview(id)
	if !ok {
		r.Response.WriteJson(g.Map{"code": 1, "msg": "not found"})
		return
	}
	gt := &tools.GerritTool{}
	if _, err := gt.PostReview(v.ChangeNum, v.Patchset, v.Payload); err != nil {
		r.Response.WriteJson(g.Map{"code": 1, "msg": "post review failed: " + err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{"code": 0, "data": g.Map{"published": true}})
}
