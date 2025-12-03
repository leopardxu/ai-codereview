package web

import (
	"eino-gerrit-review/internal/config"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func ReloadRules(r *ghttp.Request) {
	p := os.Getenv("RULE_CONFIG_PATH")
	if p == "" {
		r.Response.WriteJson(g.Map{"code": 1, "msg": "invalid path"})
		return
	}
	base := filepath.Join(".", "internal", "config")
	absBase, err := filepath.Abs(base)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": 1, "msg": "internal error"})
		return
	}
	clean := filepath.Clean(p)
	absPath, err := filepath.Abs(clean)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": 1, "msg": "invalid path"})
		return
	}
	if !strings.HasSuffix(strings.ToLower(clean), ".json") {
		r.Response.WriteJson(g.Map{"code": 1, "msg": "invalid path"})
		return
	}
	sep := string(filepath.Separator)
	if !strings.HasPrefix(absPath, absBase+sep) && absPath != absBase {
		r.Response.WriteJson(g.Map{"code": 1, "msg": "invalid path"})
		return
	}
	config.LoadRuleConfig(absPath)
	r.Response.WriteJson(g.Map{"code": 0, "msg": "ok"})
}
