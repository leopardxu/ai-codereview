package core

import (
    "context"
    "time"
)

type Executor struct{
    Timeout time.Duration
}

func (e *Executor) Run(ctx context.Context, fc *FlowContext, nodes []Node) error {
    c, cancel := context.WithTimeout(ctx, e.Timeout)
    defer cancel()
    for _, n := range nodes {
        if err := n.Run(c, fc); err != nil {
            return err
        }
    }
    return nil
}

