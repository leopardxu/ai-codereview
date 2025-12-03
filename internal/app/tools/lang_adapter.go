package tools

import (
    "regexp"
    "strings"
)

type LanguageAdapter interface{
    ExtractFunction(src string) string
    ExtractClass(src string) string
    ExtractDependencies(src string) string
}

type CAdapter struct{}
func (CAdapter) ExtractFunction(s string) string { return findFirstBlockByRegex(s, `(?m)^\s*[\w\*\[\]]+[\s\*]+[\w_]+\s*\([^\)]*\)\s*\{`) }
func (CAdapter) ExtractClass(s string) string { return "" }
func (CAdapter) ExtractDependencies(s string) string { return extractDependencies(s) }

type JavaAdapter struct{}
func (JavaAdapter) ExtractFunction(s string) string { return findFirstBlockByRegex(s, `(?m)^\s*(public|protected|private)?\s*(static\s+)?[\w\<\>\[\]]+\s+[\w_]+\s*\([^\)]*\)\s*\{`) }
func (JavaAdapter) ExtractClass(s string) string { return findFirstBlockByRegex(s, `(?m)^\s*(public|protected|private)?\s*(final|abstract)?\s*class\s+[\w_]+[^\n]*\{`) }
func (JavaAdapter) ExtractDependencies(s string) string { return extractDependencies(s) }

type KotlinAdapter struct{}
func (KotlinAdapter) ExtractFunction(s string) string { return findFirstBlockByRegex(s, `(?m)^\s*fun\s+[\w_]+\s*\([^\)]*\)\s*\{`) }
func (KotlinAdapter) ExtractClass(s string) string { return findFirstBlockByRegex(s, `(?m)^\s*(open\s+)?class\s+[\w_]+[^\n]*\{`) }
func (KotlinAdapter) ExtractDependencies(s string) string { return extractDependencies(s) }

type DefaultAdapter struct{}
func (DefaultAdapter) ExtractFunction(s string) string { return limitSize(s) }
func (DefaultAdapter) ExtractClass(s string) string { return limitSize(s) }
func (DefaultAdapter) ExtractDependencies(s string) string { return extractDependencies(s) }

func adapterForPath(p string) LanguageAdapter {
    if hasSuffix(p, ".c") || hasSuffix(p, ".h") || hasSuffix(p, ".cpp") || hasSuffix(p, ".hpp") { return CAdapter{} }
    if hasSuffix(p, ".java") { return JavaAdapter{} }
    if hasSuffix(p, ".kt") || hasSuffix(p, ".kts") { return KotlinAdapter{} }
    return DefaultAdapter{}
}

func hasSuffix(s, suf string) bool {
    n := len(s)
    m := len(suf)
    if n < m { return false }
    return s[n-m:] == suf
}

func findFirstBlockByRegex(s string, pattern string) string {
    re := regexp.MustCompile(pattern)
    loc := re.FindStringIndex(s)
    if loc == nil { return limitSize(s) }
    brace := strings.Index(s[loc[1]:], "{")
    if brace < 0 { return limitSize(s) }
    start := loc[1] + brace
    depth := 0
    end := len(s)
    for i := start; i < len(s); i++ {
        if s[i] == '{' { depth++ }
        if s[i] == '}' { depth--; if depth == 0 { end = i+1; break } }
    }
    if end <= start { return limitSize(s) }
    return s[loc[0]:end]
}
