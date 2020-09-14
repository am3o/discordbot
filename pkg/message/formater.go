package message

import (
	"fmt"
	"strings"
)

const TypeGIF = "gif"

var DefaultTextFormatter = new(QuoteFormatter)

type QuoteFormatter struct{}

func (*QuoteFormatter) Format(quote, source string) (message string) {
	switch {
	case quote == "" && source == "":
	case strings.HasPrefix(quote, "http") && strings.Contains(strings.ToLower(source), TypeGIF):
		message = quote
	default:
		message = fmt.Sprintf("> %v \n > - %v", quote, source)
	}
	return
}
