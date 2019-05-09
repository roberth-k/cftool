package pprint

import (
	"fmt"
	"github.com/fatih/color"
	"io"
	"strings"
)

var (
	Text    = color.New(color.Reset)
	Cyan    = color.New(color.FgCyan)
	Green   = color.New(color.FgGreen)
	Magenta = color.New(color.FgMagenta)
	Red     = color.New(color.FgRed)
	Yellow  = color.New(color.FgYellow)
)

var colors = []*color.Color{Cyan, Green, Magenta, Red, Yellow}

var (
	ColField      = Cyan
	ColAdd        = Green
	ColModify     = Yellow
	ColRemove     = Red
	ColLogicalId  = Magenta
	ColWarning    = Yellow
	ColError      = Red
	ColVerbose    = Yellow
	ColDiffHeader = Cyan
	ColDiffAdd    = Green
	ColDiffRemove = Red
	ColDiffText   = Text
)

func EnableColor() {
	for _, col := range colors {
		col.EnableColor()
	}
}

func DisableColor() {
	for _, col := range colors {
		col.DisableColor()
	}
}

func Promptf(w io.Writer, text string, args ...interface{}) bool {
	for {
		_, _ = fmt.Fprintf(w, text+" [y/n] ", args...)
		var input string
		_, _ = fmt.Scan(&input)

		switch input {
		case "y":
			return true

		case "n":
			return false

		default:
			_, _ = fmt.Fprintf(w, "Please answer y or n.\n")
		}
	}
}

func Errorf(w io.Writer, format string, args ...interface{}) {
	ColError.Fprintf(w, "ERROR! "+format, args...)
	fmt.Fprintf(w, "\n")
}

func Verbosef(w io.Writer, format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	str = "VERBOSE: " + strings.Replace(str, "\n", "\nVERBOSE: ", -1)
	ColVerbose.Fprintf(w, str)
	fmt.Fprintf(w, "\n")
}

func Warningf(w io.Writer, format string, args ...interface{}) {
	ColWarning.Fprintf(w, "WARNING! "+format, args...)
	fmt.Fprintf(w, "\n")
}
