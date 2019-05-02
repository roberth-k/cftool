package main

import (
	"fmt"
	"github.com/pborman/getopt/v2"
	"github.com/pkg/errors"
	"github.com/tetratom/cfn-tool/cli/internal"
	"github.com/tetratom/cfn-tool/manifest"
	"os"
	"strings"
)

type Deploy struct {
	Prog         *Program
	Yes          bool
	ManifestFile string
	Stack        string
	Tenant       string
}

func (d *Deploy) ParseArgs(args []string) []string {
	flags := getopt.New()
	flags.FlagLong(&d.Yes, "yes", 'y', "do not prompt for confirmation")
	flags.FlagLong(&d.ManifestFile, "tool-file", 'f', "tool file path")
	flags.FlagLong(&d.Stack, "stack", 's', "stack to deploy")
	flags.FlagLong(&d.Tenant, "tenant", 't', "tenant to deploy for")
	flags.Parse(args)
	rest := flags.Args()

	if len(rest) != 0 {
		fmt.Printf("Did not expect positional parameters.\n")
		os.Exit(1)
	}

	if d.ManifestFile == "" {
		d.ManifestFile = ".cfn-tool.yml"
	}

	if strings.HasPrefix(d.ManifestFile, "~/") {
		homedir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("unable to determine user home directory")
			os.Exit(1)
		}

		d.ManifestFile = homedir + d.ManifestFile[1:]
	}

	if _, err := os.Stat(d.ManifestFile); os.IsNotExist(err) {
		fmt.Printf("%s does not exist\n", d.ManifestFile)
		os.Exit(1)
	}

	return rest
}

func (prog *Program) Deploy(args []string) error {
	d := Deploy{}
	d.ParseArgs(args)

	fp, err := os.Open(d.ManifestFile)
	if err != nil {
		return errors.Wrapf(err, "open %s", d.ManifestFile)
	}

	m, err := manifest.Parse(fp)
	fp.Close()
	if err != nil {
		return errors.Wrap(err, "parse manifest")
	}

	decisions, err := m.Process(manifest.ProcessInput{
		Stack:  d.Stack,
		Tenant: d.Tenant})
	if err != nil {
		return errors.Wrap(err, "process manifest")
	}

	err = prog.Whoami([]string{})
	if err != nil {
		return errors.Wrap(err, "whoami")
	}

	engine := internal.NewEngine(prog.AWS.Session())

	for _, decision := range decisions {
		if d.Yes && !decision.Protected {
			decision.Protected = false
		}

		err = engine.Deploy(os.Stdout, decision)
		if err != nil {
			return errors.Wrap(err, "deploy stack")
		}
	}

	return nil
}
