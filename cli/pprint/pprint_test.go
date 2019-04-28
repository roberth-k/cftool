package pprint

import (
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestSmoke(t *testing.T) {
	DisableColor()
	defer EnableColor()

	b := strings.Builder{}
	SetWriter(&b)

	t.Run("Field", func(t *testing.T) {
		b.Reset()
		Field("  Greetings", "programs!")
		require.Equal(t, "  Greetings: programs!\n", b.String())
	})

	changeActionTests := []struct {
		Action string
		Symbol string
	}{
		{cf.ChangeActionRemove, "-"},
		{cf.ChangeActionModify, "~"},
		{cf.ChangeActionAdd, "+"},
	}

	for _, test := range changeActionTests {
		t.Run("PrintChangeHeader: "+test.Action, func(t *testing.T) {
			b.Reset()
			PrintChangeHeader(test.Action, "AWS::Resource", "MyResource")
			require.Equal(t, test.Symbol+" AWS::Resource MyResource\n", b.String())
		})
	}
}
