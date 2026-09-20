package typesafe

import (
	"encoding/json"
	"testing"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

func TestApplyAnswersRespectsConfidence(t *testing.T) {
	article := domain.Article{Kind: domain.ArticleKindNews}
	resp := systemOneResponse{Answers: map[string]json.RawMessage{
		questionKind:    json.RawMessage(`{"choice":"tutorial","confidence":0.2}`),
		questionLevel:   json.RawMessage(`{"score":2.1,"confidence":0.9}`),
		questionQuality: json.RawMessage(`{"score":1.4,"confidence":0.8}`),
		questionPromo:   json.RawMessage(`{"noul":0.81}`),
	}}
	applyAnswers(&article, resp, 0.45, 0.7)
	if article.Kind != domain.ArticleKindNews {
		t.Fatalf("low-confidence kind overwrote: %s", article.Kind)
	}
	if article.Level != domain.ArticleLevelAdvanced {
		t.Fatalf("level: %s", article.Level)
	}
	if article.Quality == nil || *article.Quality != 1.4 {
		t.Fatalf("quality: %#v", article.Quality)
	}
	if !article.Promo {
		t.Fatal("promo")
	}
}

func TestLevelFromScore(t *testing.T) {
	if got := levelFromScore(0.4); got != domain.ArticleLevelBeginner {
		t.Fatalf("got %s", got)
	}
	if got := levelFromScore(1.0); got != domain.ArticleLevelIntermediate {
		t.Fatalf("got %s", got)
	}
	if got := levelFromScore(1.6); got != domain.ArticleLevelAdvanced {
		t.Fatalf("got %s", got)
	}
}
