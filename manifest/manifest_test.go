package manifest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratom/cftool/pkg/cftool"
	"io/ioutil"
	"os"
	"testing"
)

func readAll(filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return data
}

func TestManifest_FindDeployment(t *testing.T) {
	tests := []struct {
		File        string
		TenantLabel string
		StackLabel  string
		Expect      *cftool.Deployment
	}{
		{
			File:        "testdata/mystack-manifest.yml",
			TenantLabel: "test",
			StackLabel:  "mystack",
			Expect: &cftool.Deployment{
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
				StackLabel:   "mystack",
				TenantLabel:  "test",
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
		{
			File:        "testdata/mystack-manifest.yml",
			TenantLabel: "live-us",
			StackLabel:  "mystack",
			Expect: &cftool.Deployment{
				AccountId: "111111111111",
				Parameters: map[string]string{
					"Foo":         "Bax",
					"Environment": "live",
					"SomeConst":   "bax",
				},
				StackName:    "live-mystack-us",
				TemplateBody: readAll("testdata/templates/mystack.yml"),
				Region:       "us-west-1",
				Protected:    true,
				StackLabel:   "mystack",
				TenantLabel:  "live-us",
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
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			f, err := os.Open(test.File)
			defer f.Close()
			require.NoError(t, err)
			m, err := Read(f)
			require.NoError(t, err)
			actual, found, err := m.FindDeployment(test.TenantLabel, test.StackLabel)
			require.NoError(t, err)
			require.True(t, found)
			assert.Equal(t, test.Expect, actual)
		})
	}
}
