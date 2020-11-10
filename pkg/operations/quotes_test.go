package operations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuotesOperation_ContainsKeyword(t *testing.T) {
	tt := []struct {
		text     string
		expected bool
	}{
		{
			text:     "",
			expected: false,
		},
		{
			text:     "!Foo ",
			expected: false,
		},
		{
			text:     "!foo ",
			expected: true,
		},
		{
			text:     "!Foobar",
			expected: false,
		},
		{
			text:     "!Fooo",
			expected: false,
		},
		{
			text:     "!oFooo",
			expected: false,
		},
		{
			text:     "lorem ipsum dolor sit amet, !foo consetetur sadipscing elitr",
			expected: true,
		},
		{
			text:     "lorem ipsum dolor sit amet, consetetur sadipscing elitr !foo",
			expected: true,
		},
		{
			text:     "!foo lorem ipsum dolor sit amet, consetetur sadipscing elitr",
			expected: true,
		},
		{
			text:     "!foo! lorem ipsum dolor sit amet, consetetur sadipscing elitr",
			expected: true,
		},
		{
			text:     "lorem ipsum dolor sit !foo! amet, consetetur sadipscing elitr",
			expected: true,
		},
		{
			text:     "!foo-",
			expected: true,
		},
		{
			text:     "!foo\t",
			expected: true,
		},
		{
			text:     "!foo\n",
			expected: true,
		},
		{
			text:     "lorem ipsum dolor sit foo amet, consetetur sadipscing elitr!\n",
			expected: false,
		},
	}

	t.Parallel()
	for _, keyword := range []string{
		"Foo",
		"foo",
		"fOo",
		"FOO",
	} {
		var quote = NewQuote(keyword, nil)
		for i, tc := range tt {
			t.Run(fmt.Sprintf("%v_%x", keyword, i), func(t *testing.T) {
				assert.Equal(t, tc.expected, quote.ContainsKeyword(tc.text))
			})
		}
	}
}
