package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuoteFormatter_Format(t *testing.T) {
	tt := []struct {
		quote        string
		source       string
		expectedText string
	}{
		{
			quote:        "",
			source:       "",
			expectedText: "",
		},
		{
			quote:        "foo",
			source:       "bar",
			expectedText: "> foo \n > - bar",
		},
		{
			quote:        "https://example.com",
			source:       "gif",
			expectedText: "https://example.com",
		},
		{
			quote:        "http://example.com",
			source:       "gif",
			expectedText: "http://example.com",
		},
		{
			quote:        "http://example.com",
			source:       "bar",
			expectedText: "> http://example.com \n > - bar",
		},
		{
			quote:        "https://example.com",
			source:       "bar",
			expectedText: "> https://example.com \n > - bar",
		},
	}

	t.Parallel()
	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			formatter := QuoteFormatter{}
			assert.Equal(t, tc.expectedText, formatter.Format(tc.quote, tc.source))
		})
	}
}
