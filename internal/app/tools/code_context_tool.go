package tools

import (
	"eino-gerrit-review/internal/app/policies"
	"eino-gerrit-review/internal/monitor"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ContextInfo struct {
	FilePath    string
	ContextType string
	Content     string
	StartLine   int
	EndLine     int
	Source      string
}

type CodeContextTool struct{}

var ctxCache sync.Map
var ctxLimiter = policies.NewRateLimiter(10)

func (t *CodeContextTool) Fetch(enable bool, changeNum, patchset string, diffs []map[string]interface{}) []ContextInfo {
	if !enable {
		return []ContextInfo{}
	}

	type result struct {
		info ContextInfo
		err  error
	}

	results := make(chan result, len(diffs))
	var wg sync.WaitGroup

	g := &GerritTool{}
	filter := &FileFilter{}

	for _, d := range diffs {
		wg.Add(1)
		go func(d map[string]interface{}) {
			defer wg.Done()

			p := d["path"].(string)

			// Skip files that should not be reviewed
			if filter.ShouldSkipFile(p) {
				return
			}

			key := p + "@rev"
			monitor.IncContextCall()

			if v, ok := ctxCache.Load(key); ok {
				item := v.(struct {
					exp time.Time
					val string
				})
				if time.Now().Before(item.exp) {
					monitor.IncContextHit()
					results <- result{info: ContextInfo{
						FilePath:    p,
						ContextType: granularity(),
						Content:     item.val,
						StartLine:   1,
						EndLine:     len(strings.Split(item.val, "\n")),
						Source:      "gerrit",
					}}
					return
				}
			}

			ctxLimiter.Acquire()
			// Fetch original file content from parent revision (base), not the modified version
			content, _ := g.GetFileContent(changeNum, patchset, p)

			// size limit in KB
			limitKB := atoi(getenv("CONTEXT_FILE_LIMIT", "10"))
			if limitKB > 0 {
				max := limitKB * 1024
				if len(content) > max {
					content = content[:max]
				}
			}

			monitor.IncContextMiss()
			ctxCache.Store(key, struct {
				exp time.Time
				val string
			}{exp: time.Now().Add(300 * time.Second), val: content})

			gr := granularity()
			ad := adapterForPath(p)
			var finalContent string

			switch gr {
			case "dependency":
				monitor.IncContextDep()
				finalContent = ad.ExtractDependencies(content)
			case "function":
				monitor.IncContextFunc()
				finalContent = ad.ExtractFunction(content)
			case "class":
				monitor.IncContextClass()
				finalContent = ad.ExtractClass(content)
			default:
				finalContent = content
			}

			results <- result{info: ContextInfo{
				FilePath:    p,
				ContextType: gr,
				Content:     finalContent,
				StartLine:   1,
				EndLine:     len(strings.Split(finalContent, "\n")),
				Source:      "gerrit",
			}}
		}(d)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	res := make([]ContextInfo, 0, len(diffs))
	for r := range results {
		if r.info.FilePath != "" {
			res = append(res, r.info)
		}
	}
	return res
}

func granularity() string {
	g := os.Getenv("CONTEXT_GRANULARITY")
	switch g {
	case "function", "class", "file", "dependency":
		return g
	default:
		return "file"
	}
}

func getenv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func extractDependencies(s string) string {
	lines := strings.Split(s, "\n")
	var out strings.Builder
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		if strings.HasPrefix(l, "#include ") || strings.HasPrefix(l, "import ") || strings.HasPrefix(l, "package ") ||
			strings.HasPrefix(l, "<manifest") || strings.HasPrefix(l, "<uses-permission") ||
			strings.HasPrefix(l, "apply plugin:") || strings.HasPrefix(l, "implementation ") {
			out.WriteString(l)
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func extractFirstFunction(s string) string {
	lines := strings.Split(s, "\n")
	var buf strings.Builder
	depth := 0
	started := false

	for _, l := range lines {
		if !started {
			if (strings.Contains(l, "(") && strings.Contains(l, ")") && strings.Contains(l, "{")) ||
				(strings.Contains(l, "(") && strings.Contains(l, ")") && !strings.Contains(l, ";") && strings.HasSuffix(strings.TrimSpace(l), "{")) {
				started = true
				depth = 1
				buf.WriteString(l)
				buf.WriteByte('\n')
				continue
			}
			continue
		}

		buf.WriteString(l)
		buf.WriteByte('\n')
		depth += strings.Count(l, "{")
		depth -= strings.Count(l, "}")
		if depth <= 0 {
			break
		}
	}

	res := buf.String()
	if res == "" {
		return limitSize(s)
	}
	return res
}

func extractFirstClass(s string) string {
	lines := strings.Split(s, "\n")
	var buf strings.Builder
	depth := 0
	started := false

	for _, l := range lines {
		if !started {
			if strings.HasPrefix(l, "class ") || strings.Contains(l, " class ") {
				started = true
				depth = 1
				buf.WriteString(l)
				buf.WriteByte('\n')
				continue
			}
			continue
		}

		buf.WriteString(l)
		buf.WriteByte('\n')
		depth += strings.Count(l, "{")
		depth -= strings.Count(l, "}")
		if depth <= 0 {
			break
		}
	}

	res := buf.String()
	if res == "" {
		return limitSize(s)
	}
	return res
}

func limitSize(s string) string {
	limit := atoi(getenv("CONTEXT_FILE_LIMIT", "10")) * 1024
	if limit <= 0 {
		return s
	}
	if len(s) > limit {
		return s[:limit]
	}
	return s
}
