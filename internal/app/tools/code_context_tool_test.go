package tools

import "testing"

func TestFetchContextBasic(t *testing.T) {
	diffs := []map[string]interface{}{{"path": "kernel/lock.c", "patch": ""}}
	out := (&CodeContextTool{}).Fetch(true, "123", "1", diffs)
	if len(out) == 0 {
		t.Fatalf("expected context info")
	}
	if out[0].FilePath != "kernel/lock.c" {
		t.Fatalf("wrong file")
	}
}
