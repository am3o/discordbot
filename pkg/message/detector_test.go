package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeywordDetector_Contains(t *testing.T) {
	tt := []struct{
		text string
		includesKeyword bool
	}{
		{
			text:            "",
			includesKeyword: false,
		},
		{
			text:            "!Foo",
			includesKeyword: true,
		},
		{
			text:            "!Foobar",
			includesKeyword: false,
		},
		{
			text:            "lorem ipsum dolor sit amet, !foo consetetur sadipscing elitr",
			includesKeyword: true,
		},
		{
			text:            "lorem ipsum dolor sit amet, consetetur sadipscing elitr !foo",
			includesKeyword: true,
		},
		{
			text:            "!foo lorem ipsum dolor sit amet, consetetur sadipscing elitr",
			includesKeyword: true,
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
				assert.Equal(t, tc.includesKeyword, detector.IsKeywordIncluded(tc.text))
			})
		}
	}
}
