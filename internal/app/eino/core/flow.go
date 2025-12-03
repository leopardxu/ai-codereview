package core

import "context"

type Result map[string]interface{}

type Flow interface {
    Execute(ctx context.Context, fc *FlowContext) (Result, error)
}

