package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeywordDetector_Contains(t *testing.T) {
	tt := []struct {
		text          string
		expectKeyword bool
	}{
		{
			text:          "",
			expectKeyword: false,
		},
		{
			text:          "!Foo",
			expectKeyword: true,
		},
		{
			text:          "!Foobar",
			expectKeyword: false,
		},
		{
			text:          "!Fooo",
			expectKeyword: false,
		},
		{
			text:          "!oFooo",
			expectKeyword: false,
		},
		{
			text:          "lorem ipsum dolor sit amet, !foo consetetur sadipscing elitr",
			expectKeyword: true,
		},
		{
			text:          "lorem ipsum dolor sit amet, consetetur sadipscing elitr !foo",
			expectKeyword: true,
		},
		{
			text:          "!foo lorem ipsum dolor sit amet, consetetur sadipscing elitr",
			expectKeyword: true,
		},
		{
			text:          "!foo! lorem ipsum dolor sit amet, consetetur sadipscing elitr",
			expectKeyword: true,
		},
		{
			text:          "lorem ipsum dolor sit!foo! amet, consetetur sadipscing elitr",
			expectKeyword: true,
		},
		{
			text:          "!foo-",
			expectKeyword: true,
		},
		{
			text:          "!foo\t",
			expectKeyword: true,
		},
		{
			text:          "!foo\n",
			expectKeyword: true,
		},
	}

	t.Parallel()
	for _, keyword := range []string{
		"Foo",
		"foo",
		"fOo",
		"FOO",
	} {
		var detector = NewKeywordDetector(keyword)
		for _, tc := range tt {
			t.Run(tc.text, func(t *testing.T) {
				assert.Equal(t, tc.expectKeyword, detector.IsKeywordIncluded(tc.text))
			})
		}
	}
}
