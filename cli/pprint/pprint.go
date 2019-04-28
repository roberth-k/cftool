package pprint

import (
	"fmt"
	"github.com/fatih/color"
	"os"
)

var printfCyan = color.New(color.FgCyan).PrintfFunc()
var fprintfRed = color.New(color.FgRed).FprintfFunc()
var printfYellow = color.New(color.FgYellow).PrintfFunc()

func stringize(value interface{}) string {
	switch x := value.(type) {
	case fmt.Stringer:
		return x.String()
	case string:
		return x
	default:
		panic("expected something stringable")
	}
}

func Field(field string, value interface{}) {
	printfCyan("%s:", field)
	fmt.Println(" " + stringize(value))
}

func Errorf(format string, args ...interface{}) {
	fprintfRed(os.Stderr, "ERROR! "+format, args...)
	_, _ = fmt.Fprintf(os.Stderr, "\n")
}

func Prompt(format string, args ...interface{}) bool {
	fmt.Printf(format+" [y/n] ", args...)
	var input string
	_, _ = fmt.Scan(&input)
	return input == "y"
}

func Write(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func UserErrorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func Verbosef(format string, args ...interface{}) {
	printfYellow("VERBOSE: "+format, args...)
	fmt.Printf("\n")
}
