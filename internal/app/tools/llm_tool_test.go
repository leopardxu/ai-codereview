package tools

import (
	"encoding/json"
	"testing"
)

func TestLLMAdviceParsing(t *testing.T) {
	// Simulate a response from the LLM
	jsonContent := `
	[
		{
			"Severity": "high",
			"Title": "Test Issue",
			"Detail": "This is a test detail",
			"Suggest": "Fix it",
			"File": "test.go",
			"Line": 10
		}
	]
	`

	var advice []LLMAdvice
	err := json.Unmarshal([]byte(jsonContent), &advice)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(advice) != 1 {
		t.Fatalf("Expected 1 advice, got %d", len(advice))
	}

	if advice[0].Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", advice[0].Title)
	}
}

func TestLLMAdviceParsing_Empty(t *testing.T) {
	jsonContent := `[]`
	var advice []LLMAdvice
	err := json.Unmarshal([]byte(jsonContent), &advice)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if len(advice) != 0 {
		t.Fatalf("Expected 0 advice, got %d", len(advice))
	}
}
