package tools

import "testing"

func TestDiffBinaryFilter(t *testing.T) {
    diffs := []map[string]interface{}{{"path": "image.png", "patch": ""}, {"path": "code.c", "patch": "line\nline"}}
    out := (&DiffTool{}).Parse(diffs)
    if len(out) != 1 || out[0]["path"].(string) != "code.c" { t.Fatalf("binary filter failed") }
}

