package manifest

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseParameterFile(t *testing.T) {
	tests := []struct {
		Input  string
		Expect map[string]string
	}{
		{
			"testdata/parameters1.json",
			map[string]string{"Foo": "Bar"},
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			actual, err := ReadParametersFromFile(test.Input)
			require.NoError(t, err)
			require.Equal(t, test.Expect, actual)
		})
	}
}
