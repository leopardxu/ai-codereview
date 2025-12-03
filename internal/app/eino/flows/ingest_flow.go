package flows

import (
    "context"
    "eino-gerrit-review/internal/app/eino/core"
)

type IngestFlow struct{}

func (f *IngestFlow) Execute(ctx context.Context, fc *core.FlowContext) (core.Result, error) {
    return core.Result{"changes": []map[string]interface{}{}}, nil
}

