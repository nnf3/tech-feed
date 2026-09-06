package domain

import "testing"

func TestTopTagsRanksAndExcludes(t *testing.T) {
	got := TopTags([]string{"Go", "go", "rust", "poem", "rust", "rust"}, 2, []string{"poem"})
	if len(got) != 2 || got[0] != "rust" || got[1] != "go" {
		t.Fatalf("%#v", got)
	}
}

func TestHistoryTagWeightsUseViewCount(t *testing.T) {
	byID := map[string]Article{
		"a": {ID: "a", Tags: []string{"go"}},
	}
	got := HistoryTagWeights(byID, []HistoryEntry{{ArticleID: "a", ViewCount: 3}})
	if len(got) != 3 {
		t.Fatalf("%#v", got)
	}
}

func TestUniqueIDs(t *testing.T) {
	got := UniqueIDs([]string{" a ", "b"}, []string{"b", ""})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("%#v", got)
	}
}
