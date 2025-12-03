package config

import "os"

type C struct {
    GerritBase   string
    GerritUser   string
    GerritToken  string
    ContextEnabled bool
    ModelName    string
    ModelMaxTokens int
    WorkerNum    int
    RateLimitQPS int
}

func Load() *C {
    return &C{
        GerritBase:     os.Getenv("GERRIT_BASE_URL"),
        GerritUser:     os.Getenv("GERRIT_USER"),
        GerritToken:    os.Getenv("GERRIT_TOKEN"),
        ContextEnabled: os.Getenv("CONTEXT_ENABLED") != "false",
        ModelName:      getenv("MODEL_NAME", "light-llm"),
        ModelMaxTokens: atoi(getenv("MODEL_MAX_TOKENS", "1024")),
        WorkerNum:      atoi(getenv("WORKER_NUM", "8")),
        RateLimitQPS:   atoi(getenv("RATE_LIMIT_QPS", "5")),
    }
}

func getenv(k, d string) string { v := os.Getenv(k); if v == "" { return d }; return v }
func atoi(s string) int { n := 0; for i:=0;i<len(s);i++{ c:=s[i]-'0'; if c>9 {break}; n = n*10 + int(c) }; return n }
