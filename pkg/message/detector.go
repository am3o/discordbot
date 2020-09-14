package message

import (
	"fmt"
	"regexp"
	"strings"
)

type KeywordDetector struct {
	Key string
	exp *regexp.Regexp
}

func NewKeywordDetector(keyword string) KeywordDetector {
	key := strings.ToLower(keyword)
	return KeywordDetector{
		Key: keyword,
		exp: regexp.MustCompile(fmt.Sprintf("!%s$|!%s\\s", key, key)),
	}
}

func (q *KeywordDetector) IsKeywordIncluded(message string) bool {
	return q.exp.MatchString(message)
}
