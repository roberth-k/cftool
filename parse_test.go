package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func checkParam(t *testing.T, input string, expectedKey string, expectedValue string) {
	k, v := parseParameterString(input)
	assert.Equal(t, expectedKey, k)
	assert.Equal(t, expectedValue, v)
}

func TestParseParameterFromCommandLine(t *testing.T) {
	checkParam(t, "a=b", "a", "b")
	checkParam(t, "a=", "a", "")
	checkParam(t, "a==b", "a", "=b")
	checkParam(t, "a==", "a", "=")
}
