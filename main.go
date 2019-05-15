package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/pborman/getopt/v2"
	"github.com/pkg/errors"
	"github.com/tetratom/cfn-tool/internal"
	"github.com/tetratom/cfn-tool/pprint"
	"os"
	"runtime"
)

var w = color.Output

var (
	gitVersion = "?.?.?"
	gitCommit  = "?????"
	progName   = "cfn-tool"
)

func versionString() string {
	return fmt.Sprintf(
		"%s %s (%s) %s %s-%s",
		progName, gitVersion, gitCommit,
		runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

type Program struct {
	AWS     internal.AWSOptions
	Verbose bool
	Color   bool
	Version bool
}

func (prog *Program) ParseFlags(argv []string) []string {
	set := getopt.New()
	set.FlagLong(&prog.AWS.Region, "region", 'r', "AWS region")
	set.FlagLong(&prog.AWS.Profile, "profile", 'p', "AWS credential profile")
	set.FlagLong(&prog.AWS.Endpoint, "endpoint", 'e', "AWS API endpoint")
	set.FlagLong(&prog.Verbose, "verbose", 'v', "enable verbose output")
	color := set.EnumLong(
		"color", 'c', []string{"on", "off"}, "on",
		"'on' or 'off'. pass 'off' to disable colors.")
	version := set.BoolLong("version", 0, "show version and exit")

	set.Parse(argv)

	if color == nil || *color == "on" {
		prog.Color = true
	}

	if version != nil {
		prog.Version = *version
	}

	return set.Args()
}

func main() {
	prog := Program{}
	rest := prog.ParseFlags(os.Args)

	if !prog.Color {
		pprint.DisableColor()
	}

	if prog.Version {
		fmt.Println(versionString())
		os.Exit(0)
	}

	if len(rest) < 1 {
		fmt.Printf("Expected a subcommand: deploy, update, or whoami.\n")
		os.Exit(1)
	}

	var err error

	switch rest[0] {
	case "deploy":
		err = prog.Deploy(rest)

	case "list":
		err = prog.List(rest)

	case "update":
		err = prog.Update(rest)

	case "whoami":
		err = prog.Whoami(rest)

	default:
		fmt.Printf("unrecognized command: %s\n", rest[0])
		os.Exit(1)
	}

	if err != nil {
		if errors.Cause(err) == internal.ErrAbortedByUser {
			fmt.Fprintf(w, "Aborted by user.\n")
			os.Exit(1)
		}

		fmt.Printf("error: %s\n", err) // TODO: %+v
		os.Exit(1)
	}

	os.Exit(0)
}

func (p *Program) Verbosef(msg string, args ...interface{}) {
	if p.Verbose {
		pprint.Verbosef(w, msg, args...)
	}
}
