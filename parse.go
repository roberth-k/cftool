package main

import (
	"strings"
)

func parseParameterString(str string) (string, string) {
	split := strings.SplitN(str, "=", 2)
	key := split[0]
	value := ""

	if len(split) > 1 {
		value = split[1]
	}

	return key, value
}
