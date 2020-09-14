package message

import (
	"fmt"
	"regexp"
)

type KeywordDetector struct {
	Key string
	exp *regexp.Regexp
}

func NewKeywordDetector(keyword string) KeywordDetector {
	return KeywordDetector{
		Key: keyword,
		exp: regexp.MustCompile(fmt.Sprintf("!%s$|!%s\\s", keyword, keyword)),
	}
}

func (q *KeywordDetector) IsKeywordIncluded(message string) bool {
	return q.exp.FindString(message) != ""
}
