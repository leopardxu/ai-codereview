package flows

import (
    "context"
    "eino-gerrit-review/internal/app/eino/core"
)

type PublishFlow struct{}

func (f *PublishFlow) Execute(ctx context.Context, fc *core.FlowContext) (core.Result, error) {
    return core.Result{"published": true}, nil
}

