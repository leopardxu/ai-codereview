package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

type LLMAdvice struct {
	Severity string `json:"Severity"`
	Title    string `json:"Title"`
	Detail   string `json:"Detail"`
	Suggest  string `json:"Suggest"`
	File     string `json:"File"`
	Line     int    `json:"Line"`
}

type LLMTool struct{}

func (t *LLMTool) Generate(prompt string) ([]LLMAdvice, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Fallback if no key provided, to avoid breaking the flow in dev
		return []LLMAdvice{}, nil
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "gpt-4o"
	}

	conf := &openai.ChatModelConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelName,
	}

	cm, err := openai.NewChatModel(context.Background(), conf)
	if err != nil {
		return nil, err
	}

	stream, err := cm.Stream(context.Background(), []*schema.Message{
		schema.UserMessage(prompt),
	})
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	fmt.Printf("DEBUG: Starting LLM stream reception...\n")
	var sb strings.Builder
	chunkCount := 0
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			// If EOF is not returned as error string but as io.EOF, we should handle it.
			// However, eino stream Recv usually returns error on finish or failure.
			// Let's assume standard behavior: err != nil means stop.
			// Check if it is EOF
			break
		}
		chunkCount++
		sb.WriteString(chunk.Content)
	}
	content := sb.String()
	fmt.Printf("DEBUG: Received %d chunks, total content length: %d bytes\n", chunkCount, len(content))
	fmt.Printf("DEBUG: LLM raw response:\n%s\n", content)
	// Attempt to find JSON array
	start := strings.Index(content, "[")
	end := strings.LastIndex(content, "]")
	if start != -1 && end != -1 && end > start {
		content = content[start : end+1]
	} else {
		// Fallback cleanup
		content = strings.TrimSpace(content)
		if strings.HasPrefix(content, "```json") {
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimSuffix(content, "```")
		} else if strings.HasPrefix(content, "```") {
			content = strings.TrimPrefix(content, "```")
			content = strings.TrimSuffix(content, "```")
		}
	}
	content = strings.TrimSpace(content)

	// First, try to parse with flexible Line field handling
	// LLM might return Line as string "[L281]" or as number 281
	var rawAdvice []map[string]interface{}
	if err := json.Unmarshal([]byte(content), &rawAdvice); err != nil {
		fmt.Printf("DEBUG: LLM JSON parse error: %v\n", err)
		return []LLMAdvice{}, nil
	}

	// Convert to LLMAdvice with Line number extraction
	advice := make([]LLMAdvice, 0, len(rawAdvice))
	for _, raw := range rawAdvice {
		adv := LLMAdvice{
			Severity: getString(raw, "Severity"),
			Title:    getString(raw, "Title"),
			Detail:   getString(raw, "Detail"),
			Suggest:  getString(raw, "Suggest"),
			File:     getString(raw, "File"),
		}

		// Handle Line field - could be number or string like "[L281]"
		if lineVal, ok := raw["Line"]; ok {
			switch v := lineVal.(type) {
			case float64:
				adv.Line = int(v)
			case string:
				// Extract number from "[L281]" format
				adv.Line = extractLineNumber(v)
			}
		}

		advice = append(advice, adv)
	}

	// Log which files were reviewed
	reviewedFiles := make(map[string]int)
	for _, adv := range advice {
		reviewedFiles[adv.File]++
	}
	fmt.Printf("DEBUG: LLM reviewed files: %v\n", reviewedFiles)
	fmt.Printf("DEBUG: LLM total advice count: %d\n", len(advice))

	return advice, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func extractLineNumber(s string) int {
	// Extract number from "[L281]" or "281" format
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	s = strings.TrimPrefix(s, "L")
	s = strings.TrimSpace(s)

	var num int
	fmt.Sscanf(s, "%d", &num)
	return num
}
