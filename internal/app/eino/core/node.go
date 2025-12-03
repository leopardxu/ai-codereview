package core

import "context"

type Node interface {
    Name() string
    Run(ctx context.Context, fc *FlowContext) error
}

