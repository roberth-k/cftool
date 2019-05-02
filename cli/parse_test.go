package main

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func checkParam(t *testing.T, input string, expectedKey string, expectedValue string) {
	actual := ParseParameterFromCommandLine(input)
	assert.Equal(t, expectedKey, *actual.ParameterKey)
	assert.Equal(t, expectedValue, *actual.ParameterValue)
}

func checkFile(t *testing.T, path string, expect map[string]string) {
	actual, err := parseParameterFile(filepath.Join("testdata", path))
	assert.NoError(t, err)
	assert.Equal(t, len(expect), len(actual))

	for k := range actual {
		assert.Equal(t, expect[k], actual[k])
	}
}

func TestParseParameterFromCommandLine(t *testing.T) {
	checkParam(t, "a=b", "a", "b")
	checkParam(t, "a=", "a", "")
	checkParam(t, "a==b", "a", "=b")
	checkParam(t, "a==", "a", "=")
}

func TestParseParameterFile(t *testing.T) {
	checkFile(t, "EmptyParameterFile.json", map[string]string{})
	checkFile(t, "ParameterFile1.json", map[string]string{"A": "B", "C": "D"})
}
