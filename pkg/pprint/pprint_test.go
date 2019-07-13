package pprint

import (
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestSmoke(t *testing.T) {
	w := &strings.Builder{}

	t.Run("Field", func(t *testing.T) {
		w.Reset()
		Field(w, "  Greetings", "programs!")
		require.Equal(t, "  Greetings: programs!\n", w.String())
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
			w.Reset()
			ChangeHeader(w, test.Action, "AWS::Resource", "MyResource")
			require.Equal(t, test.Symbol+" AWS::Resource MyResource\n", w.String())
		})
	}
}
