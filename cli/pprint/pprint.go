package pprint

import (
	"fmt"
	au "github.com/logrusorgru/aurora"
)

func Field(field string, value interface{}) {
	var str string

	switch x := value.(type) {
	case fmt.Stringer:
		str = x.String()
	case string:
		str = x
	}

	fmt.Println(au.Cyan(field+": ").String() + str)
}
