package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pborman/getopt/v2"
	"github.com/pkg/errors"
	"github.com/tetratom/cfn-tool/cli/cfn"
	"github.com/tetratom/cfn-tool/cli/pprint"
	"io/ioutil"
	"os"
	"time"
)

type Update struct {
	Prog           *Program
	Parameters     []string
	ParameterFiles []string
	Yes            bool
	StackName      string
	Template       string
}

func (u *Update) Sess() *session.Session {
	return u.Prog.AWS.Session()
}

func (update *Update) ParseFlags(args []string) []string {
	flags := getopt.New()
	flags.FlagLong(&update.Parameters, "parameter", 'P', "explicit parameters")
	flags.FlagLong(&update.ParameterFiles, "parameter-file", 'p', "path to parameter file")
	flags.FlagLong(&update.Yes, "yes", 'y', "do not prompt for stack update confirmation")
	flags.FlagLong(&update.StackName, "stack-name", 'n', "override inferrred stack name")
	flags.Parse(args)
	return flags.Args()
}

func (prog *Program) Update(args []string) error {
	update := Update{Prog: prog}
	rest := update.ParseFlags(args)

	if len(rest) != 1 {
		fmt.Printf("expected positional argument with path to template\n")
		os.Exit(1)
	}

	update.Template = rest[0]

	parameters, err := update.parseAllParameters()
	if err != nil {
		return err
	}

	err = prog.Whoami([]string{})
	if err != nil {
		return err
	}

	pprint.Field("StackName", update.StackName)
	pprint.Write("")

	update.Prog.Verbosef("Finding stack %s...", update.StackName)
	exists, err := cfn.StackExists(update.Sess(), update.StackName)
	if err != nil {
		return errors.Wrap(err, "check stack exists")
	}

	if !exists {
		ok := pprint.Prompt("Stack %s does not exist. Create?", update.StackName)
		if !ok {
			pprint.Write("Aborted by user.")
			os.Exit(1)
		}

		update.Prog.Verbosef("Creating stack %s...", update.StackName)
	}

	pprint.Verbosef("reading template %s...", update.Template)
	template, err := ioutil.ReadFile(update.Template)
	if err != nil {
		return errors.Wrapf(err, "failed to read template %s", update.Template)
	}

	pprint.Verbosef("creating change set...")
	stackUpdate, err := cfn.CreateChangeSet(
		update.Sess(),
		update.StackName,
		string(template),
		parameters)
	if err != nil {
		return errors.Wrap(err, "create change set")
	}

	pprint.Verbosef("executing change set...")
	if err := stackUpdate.Execute(); err != nil {
		return errors.Wrap(err, "execute stack update")
	}

	lastStatus := cfn.StackStatus("UNKNOWN")
	pprint.Verbosef("starting terminal status wait loop...")

	for i := 0; ; i++ {
		status, err := stackUpdate.GetStatus()
		if err != nil {
			return errors.Wrap(err, "get stack status")
		}

		if status != lastStatus {
			lastStatus, i = status, 0
			pprint.Printf("\n%s", status)
			if !status.IsTerminal() {
				pprint.Printf("...")
			}
		}

		if status.IsTerminal() {
			pprint.Printf("\n")
			break
		}

		sleepTime := 5 * time.Second

		if i < 5 {
			// Rapid updates for the first 10 seconds.
			sleepTime = 2 * time.Second
		}

		time.Sleep(sleepTime)
		pprint.Printf(".")
	}

	if lastStatus.IsFailed() {
		os.Exit(1)
	}

	return nil
}

func (update *Update) parseAllParameters() (map[string]string, error) {
	files := update.ParameterFiles
	params := update.Parameters
	result := make(map[string]string)

	for _, path := range files {
		update.Prog.Verbosef("reading parameters from %s...", path)

		paramsFromFile, err := ParseParameterFile(path)

		if err != nil {
			return nil, err
		}

		for k, v := range paramsFromFile {
			if cur, ok := result[k]; ok {
				update.Prog.Verbosef(
					"override parameter %s (current value %s) with %s",
					k, cur, v)
			}

			result[k] = v
		}
	}

	if len(update.Parameters) > 0 {
		update.Prog.Verbosef("applying command-line parameter overrides...")

		for _, paramSpec := range params {
			param := ParseParameterFromCommandLine(paramSpec)
			result[*param.ParameterKey] = *param.ParameterValue
		}
	}

	return result, nil
}
