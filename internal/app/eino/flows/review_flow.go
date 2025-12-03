package flows

import (
	"context"
	einoGraph "eino-gerrit-review/internal/app/eino"
	"eino-gerrit-review/internal/app/eino/core"
	"eino-gerrit-review/internal/monitor"
	"time"

	"github.com/cloudwego/eino/compose"
)

type ReviewFlow struct{}

func (f *ReviewFlow) Execute(ctx context.Context, fc *core.FlowContext) (core.Result, error) {
	g, err := einoGraph.BuildReviewGraph()
	if err != nil {
		return nil, err
	}
	r, err := g.Compile(ctx, compose.WithMaxRunSteps(20))
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, "enableContext", fc.EnableContext)
	start := time.Now()
	out, err := r.Invoke(ctx, map[string]any{"changeId": fc.ChangeId, "patchset": fc.Patchset, "enableContext": fc.EnableContext})
	dur := time.Since(start).Milliseconds()
	if dur > 0 {
		monitor.AddGraphExecMillis(uint64(dur))
	}
	if err != nil {
		return nil, err
	}
	monitor.IncCall()
	return core.Result{"reviewId": "R0001", "preview": out["preview"]}, nil
}
