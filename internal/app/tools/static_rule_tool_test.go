package tools

import (
    "testing"
    "eino-gerrit-review/internal/config"
)

func TestStaticRules(t *testing.T) {
    config.SetRuleSwitches(config.RuleSwitches{LinuxSpinSleep: true, AndroidUiSleep: true, AndroidWebView: true, FileTooLong: true})
    ctxs := []ContextInfo{{FilePath: "kernel/lock.c", Content: "spin_lock(&l); msleep(1); spin_unlock(&l)", ContextType: "file"}}
    adv := (&StaticRuleTool{}).Run(nil, ctxs)
    if len(adv) == 0 { t.Fatalf("expected advice") }
}
