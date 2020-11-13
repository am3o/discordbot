package operations

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

type QuotesOperator []Quote

func NewQuotesOperator(dictonary map[string][]string) QuotesOperator {
	operator := make(QuotesOperator, 0)
	for key, quotes := range dictonary {
		operator = append(operator, NewQuote(key, quotes))
	}

	return operator
}

func (operator QuotesOperator) String() (result string) {
	for _, quote := range operator {
		result += fmt.Sprintf("%v\n", quote.keyword)
	}
	return
}

func (operator QuotesOperator) Exec(message string) (quotes []string) {
	message = strings.ToLower(message)
	if !strings.Contains(message, "!") {
		return
	}

	for _, operation := range operator {
		quote, err := operation.Exec(message)
		if err != nil {
			continue
		}

		quotes = append(quotes, quote)
	}

	return
}

type Quote struct {
	regex   *regexp.Regexp
	keyword string
	quotes  []string
}

func NewQuote(keyword string, quotes []string) Quote {
	const REGEX = "(?:^|\\W)!%v(?:$|\\W)"

	key := strings.ToLower(keyword)
	return Quote{
		regex:   regexp.MustCompile(fmt.Sprintf(REGEX, key)),
		keyword: key,
		quotes:  quotes,
	}
}

func (q *Quote) ContainsKeyword(message string) bool {
	return q.regex.MatchString(message)
}

func (q *Quote) Exec(message string) (string, error) {
	if !q.ContainsKeyword(message) {
		return "", fmt.Errorf("could not find any keyword: %v", q.keyword)
	}

	return q.String(rand.Int() % len(q.quotes)), nil
}

const TypeGIF = "gif"

func (q *Quote) String(id int) string {
	quote, source := q.quotes[id], q.keyword
	switch {
	case quote == "" && source == "":
		return ""
	case strings.HasPrefix(quote, "http") && strings.Contains(strings.ToLower(source), TypeGIF):
		return quote
	default:
		return fmt.Sprintf("> %v \n > - %v", quote, source)
	}
}
