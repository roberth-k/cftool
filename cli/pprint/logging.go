package pprint

import (
	"fmt"
	"os"
)

func UserErrorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func Verbosef(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}
