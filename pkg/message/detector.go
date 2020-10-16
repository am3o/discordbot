package message

import (
	"fmt"
	"regexp"
	"strings"
)

type KeywordDetector struct {
	exp *regexp.Regexp
}

func NewKeywordDetector(keyword string) KeywordDetector {
	const REGEX = "\\b%v\\b"

	key := strings.ToLower(keyword)
	return KeywordDetector{
		exp: regexp.MustCompile(fmt.Sprintf(REGEX, key)),
	}
}

func (q *KeywordDetector) IsKeywordIncluded(message string) bool {
	return q.exp.MatchString(strings.ToLower(message))
}
