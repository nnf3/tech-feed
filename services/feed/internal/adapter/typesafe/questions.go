package typesafe

import (
	"encoding/json"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

const (
	questionKind    = "kind"
	questionLevel   = "level"
	questionQuality = "quality"
	questionPromo   = "promo"
)

var enrichQuestions = map[string]question{
	questionKind: {
		Type:         "choice",
		Instructions: "`title`・`summary`・`tags` から見て、このソフトウエアエンジニア向け記事はどれに当たるか。",
		Criteria: map[string]string{
			domain.ArticleKindTutorial: "作り方・使い方の解説、手順、ウォークスルー。",
			domain.ArticleKindNews:     "ニュース、発表のまとめ、最近の出来事のレポート。",
			domain.ArticleKindOpinion:  "意見、エッセイ、キャリア、個人の振り返り。",
			domain.ArticleKindRelease:  "リリースノート、変更履歴、バージョン発表、移行ガイド。",
			domain.ArticleKindOther:    "どれにも当てはまらない。",
		},
	},
	questionLevel: {
		Type:         "score",
		Instructions: "`title`・`summary`・`tags` から見て、この記事が想定する習熟度はどれか。",
		ScoreLevels: []string{
			"入門。前提知識はほとんど要らない。",
			"実務。日常的な開発経験を前提にする。",
			"上級。深い専門知識や内部実装の理解を前提にする。",
		},
	},
	questionQuality: {
		Type:         "score",
		Instructions: "`title`・`summary`・`tags` から見て、実務のエンジニアにとってどれくらい役立つか。",
		ScoreLevels: []string{
			"薄い、汎用的、または宣伝が主で技術的な中身が少ない。",
			"一読する価値はある。具体的な記述が少しある。",
			"情報密度が高い。具体的、独自、またはすぐ手を動かせる。",
		},
	},
	questionPromo: {
		Type:         "noul",
		Instructions: "技術記事というより、宣伝・採用・製品売り込みが主目的か。",
		NoulCriteria: &noulCriteria{
			True:  "主目的はマーケティング、採用、または製品の販売である。",
			False: "主目的は技術情報またはエンジニアリングの経験である。",
		},
	},
}

type question struct {
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria,omitempty"`
	ScoreLevels  []string          `json:"-"`
	NoulCriteria *noulCriteria     `json:"-"`
}

type noulCriteria struct {
	True  string `json:"true"`
	False string `json:"false"`
}

func (q question) MarshalJSON() ([]byte, error) {
	switch q.Type {
	case "score":
		return marshalMap(map[string]any{
			"type":         q.Type,
			"instructions": q.Instructions,
			"criteria":     q.ScoreLevels,
		})
	case "noul":
		body := map[string]any{
			"type":         q.Type,
			"instructions": q.Instructions,
		}
		if q.NoulCriteria != nil {
			body["criteria"] = q.NoulCriteria
		}
		return marshalMap(body)
	default:
		return marshalMap(map[string]any{
			"type":         q.Type,
			"instructions": q.Instructions,
			"criteria":     q.Criteria,
		})
	}
}

func marshalMap(v map[string]any) ([]byte, error) {
	return json.Marshal(v)
}
