package typesafe

import (
	"encoding/json"
	"math"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type systemOneResponse struct {
	Answers map[string]json.RawMessage `json:"answers"`
}

type choiceAnswer struct {
	Choice     string  `json:"choice"`
	Confidence float64 `json:"confidence"`
}

type scoreAnswer struct {
	Score      float64 `json:"score"`
	Confidence float64 `json:"confidence"`
}

type noulAnswer struct {
	Noul float64 `json:"noul"`
}

func applyAnswers(article *domain.Article, resp systemOneResponse, minConfidence, promoThreshold float64) {
	if raw, ok := resp.Answers[questionKind]; ok {
		var ans choiceAnswer
		if json.Unmarshal(raw, &ans) == nil && ans.Confidence >= minConfidence && validKind(ans.Choice) {
			article.Kind = ans.Choice
		}
	}
	if raw, ok := resp.Answers[questionLevel]; ok {
		var ans scoreAnswer
		if json.Unmarshal(raw, &ans) == nil && ans.Confidence >= minConfidence {
			article.Level = levelFromScore(ans.Score)
		}
	}
	if raw, ok := resp.Answers[questionQuality]; ok {
		var ans scoreAnswer
		if json.Unmarshal(raw, &ans) == nil && ans.Confidence >= minConfidence {
			q := ans.Score
			article.Quality = &q
		}
	}
	if raw, ok := resp.Answers[questionPromo]; ok {
		var ans noulAnswer
		if json.Unmarshal(raw, &ans) == nil {
			article.Promo = ans.Noul >= promoThreshold
		}
	}
}

func validKind(kind string) bool {
	switch kind {
	case domain.ArticleKindTutorial, domain.ArticleKindNews, domain.ArticleKindOpinion, domain.ArticleKindRelease, domain.ArticleKindOther:
		return true
	default:
		return false
	}
}

func levelFromScore(score float64) string {
	i := int(math.Round(score))
	if i < 0 {
		i = 0
	}
	if i > 2 {
		i = 2
	}
	return []string{domain.ArticleLevelBeginner, domain.ArticleLevelIntermediate, domain.ArticleLevelAdvanced}[i]
}
