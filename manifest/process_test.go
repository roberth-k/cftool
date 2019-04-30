package manifest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestManifest_Process(t *testing.T) {
	tv := true

	tests := []struct {
		File   string
		Input  ProcessInput
		Expect []*Decision
	}{
		{
			File: "testdata/mystack-manifest.yml",
			Input: ProcessInput{
				Stack:  "mystack",
				Tenant: "test",
			},
			Expect: []*Decision{
				{
					Tenant: "test",
					Deployment: &Deployment{
						AccountId: "222222222222",
						Parameters: []*Parameter{
							{File: "stacks/test/eu-west-1/test-mystack.json"},
							{Key: "Environment", Value: "test"},
						},
						StackName: "test-mystack",
						Template:  "templates/mystack.yml",
						Region:    "eu-west-1",
						Protected: nil,
					},
				},
			},
		},
		{
			File: "testdata/mystack-manifest.yml",
			Input: ProcessInput{
				Stack:  "mystack",
				Tenant: "live-us",
			},
			Expect: []*Decision{
				{
					Tenant: "live-us",
					Deployment: &Deployment{
						AccountId: "111111111111",
						Parameters: []*Parameter{
							{File: "stacks/live/us-west-1/live-mystack-us.json"},
							{Key: "Environment", Value: "live"},
						},
						StackName: "live-mystack-us",
						Template:  "templates/mystack.yml",
						Region:    "us-west-1",
						Protected: &tv,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			f, err := os.Open(test.File)
			require.NoError(t, err)
			m, err := Parse(f)
			require.NoError(t, err)
			actual, err := m.Process(test.Input)
			require.NoError(t, err)
			assert.Equal(t, test.Expect, actual)
		})
	}
}
