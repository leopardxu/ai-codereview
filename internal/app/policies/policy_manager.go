package policies

type PolicyManager struct {
	DiffChunkLines int
	GerritQPS      int
	ContextQPS     int
	MaxTokens      int
	MaxInputChars  int
}

func Default() *PolicyManager {
	return &PolicyManager{DiffChunkLines: 300, GerritQPS: 5, ContextQPS: 10, MaxTokens: 1024, MaxInputChars: 50000}
}
