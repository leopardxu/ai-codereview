package eino

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type MockChatModel struct {
	tools []*schema.ToolInfo
}

func (m *MockChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// Always emit a tool call named "code_review" with args from context
	args := map[string]any{
		"changeNum":     ctx.Value("changeNum"),
		"patchset":      ctx.Value("patchset"),
		"enableContext": ctx.Value("enableContext"),
	}
	b, _ := json.Marshal(args)
	return &schema.Message{
		Role:      schema.Assistant,
		ToolCalls: []schema.ToolCall{{Function: schema.FunctionCall{Name: "code_review", Arguments: string(b)}}},
	}, nil
}

func (m *MockChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	sr, sw := schema.Pipe[*schema.Message](0)
	go func() {
		defer sw.Close()
		msg, _ := m.Generate(ctx, input, opts...)
		sw.Send(msg, nil)
	}()
	return sr, nil
}

func (m *MockChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	m.tools = tools
	return m, nil
}
