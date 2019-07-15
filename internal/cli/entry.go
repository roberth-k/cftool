package cli

import (
	"context"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/tetratom/cftool/internal"
	"github.com/tetratom/cftool/pkg/pprint"
	"os"
	"runtime"
	"runtime/debug"
)

var gitVersion string

func Entry(c context.Context, args []string) error {
	options := ParseGlobalOptions(args)

	if !options.Color {
		pprint.DisableColor()
	}

	if options.Version {
		fmt.Fprintf(
			color.Output,
			"cftool %s %s %s/%s",
			version(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
		return nil
	}

	if len(options.remainingArgs) < 1 {
		flag.Usage()
		fmt.Fprintf(color.Output, "\nExpected subcommand: deploy, update\n")
		os.Exit(1) // TODO: Return error instead?
	}

	var err error
	switch subcommand := options.remainingArgs[0]; subcommand {
	case "deploy":
		err = Deploy(c, options, ParseDeployOptions(options.remainingArgs))
	case "update":
		err = Update(c, options, ParseUpdateOptions(options.remainingArgs))
	default:
		// todo: where to output to?
		fmt.Fprintf(color.Output, "\nUnrecognized subcommand: %s\n", subcommand)
	}

	if err != nil {
		if errors.Cause(err) == internal.ErrAbortedByUser {
			fmt.Fprintf(color.Output, "Aborted by user.\n")
			os.Exit(1)
		}

		return err
	}

	return nil
}

func version() string {
	if gitVersion != "" {
		return gitVersion
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		return buildInfo.Main.Version
	}

	return "(unknown)"
}
