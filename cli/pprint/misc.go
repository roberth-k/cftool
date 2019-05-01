package pprint

import (
	"fmt"
	"io"
)

func BeginField(w io.Writer, field string) {
	ColField.Fprintf(w, "% 10s:", field)
	fmt.Fprintf(w, " ")
}

func Field(w io.Writer, field string, value interface{}) {
	BeginField(w, field)
	fmt.Fprintf(w, "%s\n", value)
}
