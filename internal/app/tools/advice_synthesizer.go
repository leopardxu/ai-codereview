package tools

func Synthesize(static []RuleAdvice, llm []LLMAdvice) []map[string]interface{} {
    out := make([]map[string]interface{}, 0, len(static)+len(llm))
    for _, s := range static {
        out = append(out, map[string]interface{}{
            "file":    s.File,
            "line":    s.Line,
            "severity": s.Severity,
            "message":  s.Title + ": " + s.Detail + " 建议：" + s.Suggest,
        })
    }
    for _, a := range llm {
        out = append(out, map[string]interface{}{
            "file":    a.File,
            "line":    a.Line,
            "severity": a.Severity,
            "message":  a.Title + ": " + a.Detail + " 建议：" + a.Suggest,
        })
    }
    return out
}
