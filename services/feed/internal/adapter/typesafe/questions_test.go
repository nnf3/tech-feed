package typesafe

import (
	"encoding/json"
	"testing"
)

func TestQuestionJSONShape(t *testing.T) {
	raw, err := json.Marshal(enrichQuestions[questionLevel])
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	criteria, ok := body["criteria"].([]any)
	if !ok || len(criteria) != 3 {
		t.Fatalf("score criteria: %s", raw)
	}
}
