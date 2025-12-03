package eino

import (
	"context"
	"os"
	"testing"

	"github.com/cloudwego/eino/compose"
)

func TestBuildReviewGraph(t *testing.T) {
	os.Setenv("GERRIT_BASE_URL", "")
	g, err := BuildReviewGraph()
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	r, err := g.Compile(context.Background(), compose.WithMaxRunSteps(20))
	if err != nil {
		t.Fatalf("compile err: %v", err)
	}
	out, err := r.Invoke(context.Background(), map[string]any{"changeId": "C456", "patchset": "1", "enableContext": true})
	if err != nil {
		t.Fatalf("invoke err: %v", err)
	}
	v, ok := out["preview"].(map[string]any)
	if !ok {
		t.Fatalf("preview missing")
	}
	if _, ok := v["comments"]; !ok {
		t.Fatalf("comments missing")
	}
}

func TestBuildReactGraph(t *testing.T) {
	os.Setenv("GERRIT_BASE_URL", "")
	g, err := BuildReactGraph()
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	r, err := g.Compile(context.Background(), compose.WithMaxRunSteps(20))
	if err != nil {
		t.Fatalf("compile err: %v", err)
	}
	out, err := r.Invoke(context.Background(), map[string]any{"changeId": "C456", "patchset": "1"})
	if err != nil {
		t.Fatalf("invoke err: %v", err)
	}
	if _, ok := out["preview"].(map[string]any); !ok {
		t.Fatalf("preview missing")
	}
}
