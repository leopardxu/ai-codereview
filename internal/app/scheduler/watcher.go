package scheduler

import (
	"context"
	"eino-gerrit-review/internal/app/tools"
	"fmt"
	"time"
)

type Watcher struct {
	Ticker *time.Ticker
}

func NewWatcher(d time.Duration) *Watcher { return &Watcher{Ticker: time.NewTicker(d)} }

func (w *Watcher) Run(ctx context.Context, project, branch string, pool *WorkerPool, enableContext bool) {
	gt := &tools.GerritTool{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.Ticker.C:
			changes, _ := gt.GetOpenChanges(project, branch, 10)
			for _, c := range changes {
				// Use _number field from Gerrit API response as the unique identifier
				num := ""
				if n, ok := c["_number"].(float64); ok {
					num = fmt.Sprintf("%.0f", n)
				}
				if num != "" {
					pool.Submit(Task{ChangeNum: num, Patchset: "1", EnableContext: enableContext})
				}
			}
		}
	}
}
