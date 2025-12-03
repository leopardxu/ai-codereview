package tools

import (
	"eino-gerrit-review/internal/app/policies"
	"strings"
)

type DiffTool struct{}

func (t *DiffTool) Parse(diffs []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(diffs))
	filter := &FileFilter{}

	for _, d := range diffs {
		p := d["path"].(string)
		if filter.ShouldSkipFile(p) {
			continue
		}
		lang := detectLang(p)
		patch := d["patch"].(string)
		limit := policies.Default().DiffChunkLines
		if strings.Count(patch, "\n") > limit {
			patch = truncateLines(patch, limit)
		}
		out = append(out, map[string]interface{}{"path": p, "lang": lang, "patch": patch})
	}
	return out
}

func truncateLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n")
}
