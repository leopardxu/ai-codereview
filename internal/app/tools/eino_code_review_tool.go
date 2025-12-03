package tools

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

type reviewReq struct {
	ChangeId      string `json:"changeId"`
	Patchset      string `json:"patchset"`
	EnableContext bool   `json:"enableContext"`
}

type reviewResp struct {
	Preview map[string]interface{} `json:"preview"`
}

func NewCodeReviewTool() tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "code_review",
			Desc: "Run Gerrit change review and return structured preview payload",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"changeId":      {Type: "string", Desc: "Gerrit change id"},
				"patchset":      {Type: "string", Desc: "Gerrit patchset/revision"},
				"enableContext": {Type: "boolean", Desc: "Enable context-enhanced review"},
			}),
		},
		func(ctx context.Context, in *reviewReq) (out *reviewResp, err error) {
			gt := &GerritTool{}
			diffs, _ := gt.GetDiffs(in.ChangeId, in.Patchset)
			parsed := (&DiffTool{}).Parse(diffs)
			ctxs := (&CodeContextTool{}).Fetch(in.EnableContext, in.ChangeId, in.Patchset, parsed)
			static := (&StaticRuleTool{}).Run(parsed, ctxs)
			prompt := BuildPrompt(joinPatches(parsed), ctxs)
			llm, _ := (&LLMTool{}).Generate(prompt)
			merged := Synthesize(static, llm)
			payload := FormatForGerrit(merged)
			return &reviewResp{Preview: payload}, nil
		},
	)
}

func joinPatches(diffs []map[string]interface{}) string {
	s := ""
	for _, d := range diffs {
		if v, ok := d["patch"].(string); ok {
			if p, ok := d["path"].(string); ok {
				s += "File: " + p + "\n"
			}
			s += v + "\n"
		}
	}
	return s
}

func ToolMessagesToPreview(msgs []*schema.Message) (map[string]any, error) {
	for _, m := range msgs {
		if m.Role == schema.Tool && m.Content != "" {
			var v struct {
				Preview map[string]any `json:"preview"`
			}
			if json.Unmarshal([]byte(m.Content), &v) == nil {
				return v.Preview, nil
			}
		}
	}
	return map[string]any{"message": "no preview"}, nil
}
