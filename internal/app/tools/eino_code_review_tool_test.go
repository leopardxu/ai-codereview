package tools

import (
    "testing"
    "github.com/cloudwego/eino/schema"
)

func TestToolMessagesToPreview(t *testing.T) {
    msg := &schema.Message{Role: schema.Tool, Content: `{"preview":{"message":"ok","comments":[]}}`}
    prev, err := ToolMessagesToPreview([]*schema.Message{msg})
    if err != nil { t.Fatalf("parse err: %v", err) }
    if prev["message"] == nil { t.Fatalf("missing message") }
}
