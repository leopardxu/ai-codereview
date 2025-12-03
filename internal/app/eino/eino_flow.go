package eino

import (
	"context"
	"eino-gerrit-review/internal/app/tools"
	"fmt"
	"os"

	openai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// BuildReviewGraph constructs an Eino Graph that orchestrates the review pipeline.
// Input: map[string]any{"changeId":string, "patchset":string, "enableContext":bool}
// Output: map[string]any{"preview": map[string]any}
func BuildReviewGraph() (*compose.Graph[map[string]any, map[string]any], error) {
	g := compose.NewGraph[map[string]any, map[string]any]()

	if err := g.AddLambdaNode("diff", compose.InvokableLambda(diffNode)); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode("context", compose.InvokableLambda(contextNode)); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode("analyze", compose.InvokableLambda(analyzeNode)); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode("merge", compose.InvokableLambda(mergeNode)); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode("format", compose.InvokableLambda(formatNode)); err != nil {
		return nil, err
	}

	if err := g.AddEdge(compose.START, "diff"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("diff", "context"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("context", "analyze"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("analyze", "merge"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("merge", "format"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("format", compose.END); err != nil {
		return nil, err
	}

	return g, nil
}

type DiffOutput struct {
	Diffs    []map[string]interface{}
	ChangeId string
	Patchset string
}

func diffNode(ctx context.Context, in map[string]any) (*DiffOutput, error) {
	changeId, _ := in["changeId"].(string)
	patchset, _ := in["patchset"].(string)
	fmt.Printf("DEBUG: Fetching diffs for ChangeId: %s, Patchset: %s\n", changeId, patchset)
	gt := &tools.GerritTool{}
	diffs, err := gt.GetDiffs(changeId, patchset)
	if err != nil {
		fmt.Printf("DEBUG: GetDiffs error: %v\n", err)
		return nil, err
	}
	fmt.Printf("DEBUG: Got %d raw diffs\n", len(diffs))
	out := (&tools.DiffTool{}).Parse(diffs)
	fmt.Printf("DEBUG: Parsed %d diffs\n", len(out))
	return &DiffOutput{Diffs: out, ChangeId: changeId, Patchset: patchset}, nil
}

func contextNode(ctx context.Context, in *DiffOutput) (struct {
	Diffs []map[string]interface{}
	Ctxs  []tools.ContextInfo
}, error) {
	enable, _ := ctx.Value("enableContext").(bool)
	fmt.Printf("DEBUG: Fetching context (enable=%v)\n", enable)
	ctxs := (&tools.CodeContextTool{}).Fetch(enable, in.ChangeId, in.Patchset, in.Diffs)
	fmt.Printf("DEBUG: Fetched %d context items\n", len(ctxs))
	return struct {
		Diffs []map[string]interface{}
		Ctxs  []tools.ContextInfo
	}{Diffs: in.Diffs, Ctxs: ctxs}, nil
}

func analyzeNode(ctx context.Context, in struct {
	Diffs []map[string]interface{}
	Ctxs  []tools.ContextInfo
}) (struct {
	Static []tools.RuleAdvice
	Llm    []tools.LLMAdvice
}, error) {
	fmt.Println("DEBUG: Starting analysis...")
	static := (&tools.StaticRuleTool{}).Run(in.Diffs, in.Ctxs)
	fmt.Printf("DEBUG: Static analysis found %d issues\n", len(static))
	prompt := tools.BuildPrompt(joinPatches(in.Diffs), in.Ctxs)
	fmt.Printf("DEBUG: Generated prompt size: %d bytes\n", len(prompt))
	llm, err := (&tools.LLMTool{}).Generate(prompt)
	if err != nil {
		fmt.Printf("DEBUG: LLM error: %v\n", err)
	}
	fmt.Printf("DEBUG: LLM found %d issues\n", len(llm))
	return struct {
		Static []tools.RuleAdvice
		Llm    []tools.LLMAdvice
	}{Static: static, Llm: llm}, nil
}

func mergeNode(ctx context.Context, in struct {
	Static []tools.RuleAdvice
	Llm    []tools.LLMAdvice
}) ([]map[string]interface{}, error) {
	m := tools.Synthesize(in.Static, in.Llm)
	return m, nil
}

func formatNode(ctx context.Context, advs []map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"preview": tools.FormatForGerrit(advs)}, nil
}

// joinPatches helper used in LLM prompt building
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

// BuildReactGraph demonstrates ChatTemplate + ChatModel + ToolsNode orchestration
func BuildReactGraph() (*compose.Graph[map[string]any, map[string]any], error) {
	g := compose.NewGraph[map[string]any, map[string]any]()

	// ChatTemplate
	tpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是代码评审智能体，会调用 code_review 工具生成结构化建议"),
		schema.UserMessage("评审变更 {changeId} 的补丁 {patchset}"),
	)

	// ChatModel with tools binding (OpenAI if available, otherwise Mock)
	var cmModel model.BaseChatModel
	if os.Getenv("OPENAI_API_KEY") != "" {
		ctx := context.Background()
		conf := &openai.ChatModelConfig{BaseURL: os.Getenv("OPENAI_BASE_URL"), APIKey: os.Getenv("OPENAI_API_KEY"), Model: os.Getenv("MODEL_NAME")}
		real, err := openai.NewChatModel(ctx, conf)
		if err == nil {
			info, _ := tools.NewCodeReviewTool().Info(ctx)
			tcm, _ := real.WithTools([]*schema.ToolInfo{info})
			cmModel = tcm
		}
	}
	if cmModel == nil {
		cm := &MockChatModel{}
		info := &schema.ToolInfo{Name: "code_review"}
		_, _ = cm.WithTools([]*schema.ToolInfo{info})
		cmModel = cm
	}

	tn, _ := compose.NewToolNode(context.Background(), &compose.ToolsNodeConfig{Tools: []tool.BaseTool{tools.NewCodeReviewTool()}})
	if err := g.AddChatTemplateNode("template", tpl); err != nil {
		return nil, err
	}
	if err := g.AddChatModelNode("model", cmModel); err != nil {
		return nil, err
	}
	if err := g.AddToolsNode("tools", tn); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode("convert", compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) (map[string]any, error) {
		prev, _ := tools.ToolMessagesToPreview(msgs)
		return map[string]any{"preview": prev}, nil
	})); err != nil {
		return nil, err
	}

	if err := g.AddEdge(compose.START, "template"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("template", "model"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("model", "tools"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("tools", "convert"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("convert", compose.END); err != nil {
		return nil, err
	}

	return g, nil
}
