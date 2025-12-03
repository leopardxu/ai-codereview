package tools

import (
	"eino-gerrit-review/internal/config"
	"strings"
)

type RuleAdvice struct {
	Severity string
	Title    string
	Detail   string
	Suggest  string
	File     string
	Line     int
}

type StaticRuleTool struct{}

func (t *StaticRuleTool) Run(diffs []map[string]interface{}, ctxs []ContextInfo) []RuleAdvice {
	var out []RuleAdvice
	cfg := config.GetRuleSwitches()

	for _, c := range ctxs {
		if shouldSkip(c, cfg) {
			continue
		}

		// Linux Kernel Rules
		if c.FilePath == "kernel/lock.c" {
			if cfg.LinuxSpinSleep && strings.Contains(c.Content, "spin_lock") && strings.Contains(c.Content, "msleep") {
				out = append(out, RuleAdvice{
					Severity: "high",
					Title:    "自旋锁内睡眠",
					Detail:   "spin_lock 区间包含 msleep 可能导致死锁或调度问题",
					Suggest:  "避免在自旋锁持有期间睡眠，改用合适的同步原语或重构逻辑",
					File:     c.FilePath,
					Line:     findLine(c.Content, "msleep"),
				})
			}
		}

		// Android Rules
		if c.FilePath == "app/src/main/java/com/example/MainActivity.java" {
			if cfg.AndroidUiSleep && strings.Contains(c.Content, "Thread.sleep") {
				out = append(out, RuleAdvice{
					Severity: "high",
					Title:    "主线程阻塞",
					Detail:   "MainActivity 中调用 Thread.sleep 阻塞 UI 线程",
					Suggest:  "在后台线程执行耗时操作或使用 Handler/Post 延迟",
					File:     c.FilePath,
					Line:     findLine(c.Content, "Thread.sleep"),
				})
			}
			if cfg.AndroidWebView && strings.Contains(c.Content, "WebView") && !strings.Contains(c.Content, "setJavaScriptEnabled(false)") {
				out = append(out, RuleAdvice{
					Severity: "medium",
					Title:    "WebView 安全设置缺失",
					Detail:   "未显式关闭或管控 JavaScript，可能存在风险",
					Suggest:  "根据业务需要配置 WebSettings 并限制敏感能力",
					File:     c.FilePath,
					Line:     1,
				})
			}
		}

		// General Rules
		lines := strings.Count(c.Content, "\n") + 1
		limit := cfg.FunctionLengthLimit
		if v, ok := cfg.LengthLimitByLang[detectLangByPath(c.FilePath)]; ok && v > 0 {
			limit = v
		}
		for p, v := range cfg.PathLengthLimit {
			if v > 0 && strings.Contains(c.FilePath, p) {
				limit = v
			}
		}

		if cfg.FileTooLong && lines > limit {
			out = append(out, RuleAdvice{
				Severity: "medium",
				Title:    "文件过长",
				Detail:   "上下文内容超过限制，建议拆分以提升可维护性",
				Suggest:  "重构为更小的模块或函数",
				File:     c.FilePath,
				Line:     1,
			})
		}
	}
	return out
}

func shouldSkip(c ContextInfo, cfg config.RuleSwitches) bool {
	for _, w := range cfg.WhiteListFiles {
		if strings.Contains(c.FilePath, w) {
			return true
		}
	}
	if len(cfg.WhiteListFunctions) > 0 {
		for _, fn := range cfg.WhiteListFunctions {
			if fn != "" && strings.Contains(c.Content, fn) {
				return true
			}
		}
	}
	if len(cfg.WhiteListFilesByLang) > 0 {
		lang := detectLangByPath(c.FilePath)
		if arr, ok := cfg.WhiteListFilesByLang[lang]; ok {
			for _, w := range arr {
				if w != "" && strings.Contains(c.FilePath, w) {
					return true
				}
			}
		}
	}
	if len(cfg.WhiteListFunctionsByLang) > 0 {
		lang := detectLangByPath(c.FilePath)
		if arr, ok := cfg.WhiteListFunctionsByLang[lang]; ok {
			for _, fn := range arr {
				if fn != "" && strings.Contains(c.Content, fn) {
					return true
				}
			}
		}
	}
	return false
}

func findLine(s, sub string) int {
	idx := strings.Index(s, sub)
	if idx < 0 {
		return 1
	}
	return strings.Count(s[:idx], "\n") + 1
}

func detectLangByPath(p string) string {
	if strings.HasSuffix(p, ".c") || strings.HasSuffix(p, ".h") {
		return "c"
	}
	if strings.HasSuffix(p, ".cpp") || strings.HasSuffix(p, ".hpp") {
		return "cpp"
	}
	if strings.HasSuffix(p, ".java") {
		return "java"
	}
	if strings.HasSuffix(p, ".kt") || strings.HasSuffix(p, ".kts") {
		return "kotlin"
	}
	return "text"
}
