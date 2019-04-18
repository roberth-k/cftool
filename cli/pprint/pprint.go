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
	_, _ = fmt.Fprintf(os.Stderr, au.Red("Error! "+format+"\n").String(), args...)
}
