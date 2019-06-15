package manifest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func readAll(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func TestManifest_Process(t *testing.T) {
	tests := []struct {
		File   string
		Input  ProcessInput
		Expect []*Deployment
	}{
		{
			File: "testdata/mystack-manifest.yml",
			Input: ProcessInput{
				Stack:  "mystack",
				Tenant: "test",
			},
			Expect: []*Deployment{
				{
					AccountId: "222222222222",
					Parameters: map[string]string{
						"Foo":         "Bar",
						"Environment": "test",
						"SomeConst":   "const",
					},
					StackName:    "test-mystack",
					TemplateBody: readAll("testdata/templates/mystack.yml"),
					Region:       "eu-west-1",
					Protected:    false,
					TenantName:   "test",
					Tags: map[string]string{
						"Env": "test",
						"Bar": "const",
					},
					Constants: map[string]string{
						"LiveAccountId": "111111111111",
						"TestAccountId": "222222222222",
						"Some":          "const",
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
			Expect: []*Deployment{
				{
					AccountId: "111111111111",
					Parameters: map[string]string{
						"Foo":         "Bax",
						"Environment": "live",
						"SomeConst":   "bax",
					},
					StackName:    "live-mystack-us",
					TemplateBody: readAll("testdata/templates/mystack.yml"),
					Region:       "us-west-1",
					Protected:    false,
					TenantName:   "live-us",
					Tags: map[string]string{
						"Env": "live",
						"Bar": "bax",
					},
					Constants: map[string]string{
						"LiveAccountId": "111111111111",
						"TestAccountId": "222222222222",
						"Some":          "bax",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			f, err := os.Open(test.File)
			defer f.Close()
			require.NoError(t, err)
			m, err := Parse(f)
			require.NoError(t, err)
			actual, err := m.Process(test.Input)
			require.NoError(t, err)
			assert.Equal(t, test.Expect, actual)
		})
	}
}
