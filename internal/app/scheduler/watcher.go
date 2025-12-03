package scheduler

import (
	"context"
	"eino-gerrit-review/internal/app/tools"
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
				id := c["id"].(string)
				pool.Submit(Task{ChangeId: id, Patchset: "1", EnableContext: enableContext})
			}
		}
	}
}
