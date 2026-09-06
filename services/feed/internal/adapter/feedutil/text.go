package feedutil

import (
	"html"
	"regexp"
	"strings"
)

var htmlTag = regexp.MustCompile(`<[^>]+>`)

// NormalizeTags は前後空白を落とし、小文字化して重複を除く。
func NormalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

// Summarize は HTML / 改行を除き、180 文字に切る。
func Summarize(raw string) string {
	text := htmlTag.ReplaceAllString(raw, " ")
	text = html.UnescapeString(text)
	text = strings.Join(strings.Fields(text), " ")
	if len([]rune(text)) > 180 {
		return string([]rune(text)[:180]) + "…"
	}
	return text
}
