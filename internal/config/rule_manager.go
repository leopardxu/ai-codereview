package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type RuleSwitches struct {
	LinuxSpinSleep           bool
	AndroidUiSleep           bool
	AndroidWebView           bool
	FileTooLong              bool
	WhiteListFiles           []string
	FunctionLengthLimit      int
	WhiteListFunctions       []string
	WhiteListFilesByLang     map[string][]string
	WhiteListFunctionsByLang map[string][]string
	LengthLimitByLang        map[string]int
	PathLengthLimit          map[string]int
}

var (
	ruleCfg     RuleSwitches
	mu          sync.RWMutex
	lastModTime time.Time
)

func LoadRuleConfig(path string) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	if fi.ModTime() == lastModTime {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to read rule config: %v", err)
		return
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		log.Printf("failed to unmarshal rule config: %v", err)
		return
	}
	allow := map[string]struct{}{
		"LinuxSpinSleep":           {},
		"AndroidUiSleep":           {},
		"AndroidWebView":           {},
		"FileTooLong":              {},
		"WhiteListFiles":           {},
		"FunctionLengthLimit":      {},
		"WhiteListFunctions":       {},
		"WhiteListFilesByLang":     {},
		"WhiteListFunctionsByLang": {},
		"LengthLimitByLang":        {},
		"PathLengthLimit":          {},
	}
	for k := range raw {
		if _, ok := allow[k]; !ok {
			return
		}
	}
	var tmp RuleSwitches
	if json.Unmarshal(b, &tmp) != nil {
		return
	}
	if tmp.FunctionLengthLimit == 0 {
		tmp.FunctionLengthLimit = 200
	}
	tmp.WhiteListFiles = dedup(tmp.WhiteListFiles)
	tmp.WhiteListFunctions = dedup(tmp.WhiteListFunctions)
	mu.Lock()
	ruleCfg = tmp
	lastModTime = fi.ModTime()
	mu.Unlock()
}

func GetRuleSwitches() RuleSwitches { mu.RLock(); defer mu.RUnlock(); return ruleCfg }

func SetRuleSwitches(rs RuleSwitches) { mu.Lock(); ruleCfg = rs; mu.Unlock() }

func dedup(arr []string) []string {
	if len(arr) == 0 {
		return arr
	}
	m := make(map[string]struct{}, len(arr))
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
