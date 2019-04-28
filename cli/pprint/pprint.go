package pprint

import (
	"fmt"
	au "github.com/logrusorgru/aurora"
	"os"
)

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
	fmt.Println(au.Cyan(field+": ").String() + stringize(value))
}

func Errorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, au.Red("Error! "+format).String()+"\n", args...)
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
	_, _ = fmt.Fprintf(os.Stdout, au.Yellow("VERBOSE: "+format).String()+"\n", args...)
}
