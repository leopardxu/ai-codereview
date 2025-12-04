package scheduler

import (
	"context"
	"eino-gerrit-review/internal/app/eino/core"
	"eino-gerrit-review/internal/app/eino/flows"
)

type Task struct {
	ChangeNum     string
	Patchset      string
	EnableContext bool
}

type WorkerPool struct {
	ch chan Task
}

func NewWorkerPool(size int) *WorkerPool { return &WorkerPool{ch: make(chan Task, size)} }

func (p *WorkerPool) Submit(t Task) { p.ch <- t }

func (p *WorkerPool) Run(ctx context.Context) {
	go func() {
		f := &flows.ReviewFlow{}
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-p.ch:
				fc := core.NewFlowContext()
				fc.ChangeNum = t.ChangeNum
				fc.Patchset = t.Patchset
				fc.EnableContext = t.EnableContext

				// Use a derived context or the pool's context if needed,
				// but here we start a new background context for the task
				// to ensure it completes even if the pool shuts down immediately (optional design choice),
				// OR use the pool's context to support cancellation.
				// Let's use the pool's context to allow graceful shutdown cancellation.
				res, err := f.Execute(ctx, fc)
				if err != nil {
					// In a real app, use a logger
					// log.Printf("Task failed for %s/%s: %v", t.ChangeNum, t.Patchset, err)
					continue
				}
				if v, ok := res["preview"].(map[string]interface{}); ok {
					core.PutReview("S-"+t.ChangeNum+"-"+t.Patchset, v, t.ChangeNum, t.Patchset)
				}
			}
		}
	}()
}
