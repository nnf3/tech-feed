package domain

import (
	"errors"
	"testing"
)

func TestNormalizeFeedSort(t *testing.T) {
	if got := NormalizeFeedSort(""); got != SortNew {
		t.Fatalf("empty: %q", got)
	}
	if got := NormalizeFeedSort("recommended"); got != SortRecommended {
		t.Fatalf("recommended: %q", got)
	}
	if got := NormalizeFeedSort("RECOMMENDED"); got != SortRecommended {
		t.Fatalf("case: %q", got)
	}
	if got := NormalizeFeedSort("popular"); got != SortNew {
		t.Fatalf("unknown: %q", got)
	}
}

func TestSearchAfterRoundTrip(t *testing.T) {
	cur, err := EncodeSearchAfter([]any{1.71e12, "abc"})
	if err != nil || cur == "" {
		t.Fatalf("encode: %q %v", cur, err)
	}
	got, err := DecodeSearchAfter(cur)
	if err != nil || len(got) != 2 || got[1] != "abc" {
		t.Fatalf("decode: %#v %v", got, err)
	}
}

func TestDecodeSearchAfterRejects(t *testing.T) {
	if _, err := DecodeSearchAfter("%%%"); !errors.Is(err, ErrInvalidSearchAfter) {
		t.Fatalf("bad b64: %v", err)
	}
	if _, err := DecodeSearchAfter("e30"); !errors.Is(err, ErrInvalidSearchAfter) {
		t.Fatalf("empty object: %v", err)
	}
	got, err := DecodeSearchAfter("  ")
	if err != nil || got != nil {
		t.Fatalf("blank: %#v %v", got, err)
	}
}
