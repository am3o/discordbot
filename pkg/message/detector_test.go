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
			text:            "Lorem ipsum dolor sit amet, !Foo consetetur sadipscing elitr",
			includesKeyword: true,
		},
		{
			text:            "Lorem ipsum dolor sit amet, consetetur sadipscing elitr !Foo",
			includesKeyword: true,
		},
		{
			text:            "!Foo Lorem ipsum dolor sit amet, consetetur sadipscing elitr",
			includesKeyword: true,
		},
	}

	t.Parallel()
	var detector = NewKeywordDetector("Foo")
	for _, tc := range tt {
		t.Run(tc.text, func(t *testing.T) {
			assert.Equal(t, tc.includesKeyword,detector.IsKeywordIncluded(tc.text))
		})
	}
}
