package pprint

import (
	"fmt"
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/fatih/color"
	"io"
	"os"
	"strings"
)

var cyan = color.New(color.FgCyan)
var red = color.New(color.FgRed)
var yellow = color.New(color.FgYellow)
var green = color.New(color.FgGreen)
var magenta = color.New(color.FgMagenta)
var fp io.Writer = os.Stdout

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

func DisableColor() {
	cyan.DisableColor()
	red.DisableColor()
	yellow.DisableColor()
	green.DisableColor()
	magenta.DisableColor()
}

func EnableColor() {
	cyan.EnableColor()
	red.EnableColor()
	yellow.EnableColor()
	green.EnableColor()
	magenta.EnableColor()
}

func SetWriter(w io.Writer) {
	fp = w
}

func ResetWriter() {
	fp = os.Stdout
}

func StartField(field string) {
	_, _ = cyan.Fprintf(fp, "%s:", field)
	_, _ = fmt.Fprintf(fp, " ")
}

func Field(field string, value interface{}) {
	StartField(field)
	_, _ = fmt.Fprintln(fp, stringize(value))
}

func Errorf(format string, args ...interface{}) {
	_, _ = red.Fprintf(fp, "ERROR! "+format, args...)
	_, _ = fmt.Fprintf(fp, "\n")
}

func Prompt(format string, args ...interface{}) bool {
	_, _ = fmt.Fprintf(fp, format+" [y/n] ", args...)
	var input string
	_, _ = fmt.Scan(&input)
	return input == "y"
}

func Write(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(fp, format+"\n", args...)
}

func Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(fp, format, args...)
}

func UserErrorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(fp, format+"\n", args...)
}

func Verbosef(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	str = "VERBOSE: " + strings.Replace(str, "\n", "\nVERBOSE: ", -1)
	_, _ = yellow.Fprintf(fp, str)
	_, _ = fmt.Fprintf(fp, "\n")
}

func Warningf(format string, args ...interface{}) {
	_, _ = yellow.Fprintf(fp, "WARNING! "+format, args...)
	_, _ = fmt.Fprintf(fp, "\n")
}

func PrintChangeHeader(action string, resourceType string, logicalResourceId string) {
	symbol := "???"
	col := yellow

	switch action {
	case cf.ChangeActionRemove:
		symbol = "-"
		col = red

	case cf.ChangeActionModify:
		symbol = "~"
		col = yellow

	case cf.ChangeActionAdd:
		symbol = "+"
		col = green
	}

	_, _ = col.Fprintf(fp, "%s %s", symbol, resourceType)
	_, _ = magenta.Fprintf(fp, " %s", logicalResourceId)
	_, _ = fmt.Fprintln(fp)
}

func Yellowf(format string, args ...interface{}) {
	_, _ = yellow.Fprintf(fp, format, args...)
}

func Redf(format string, args ...interface{}) {
	_, _ = red.Fprintf(fp, format, args...)
}
