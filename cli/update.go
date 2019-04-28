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

func (prog *Program) Update(args []string) error {
	update := Update{Prog: prog}

	flags := getopt.New()
	flags.FlagLong(&update.Parameters, "parameter", 'P', "explicit parameters")
	flags.FlagLong(&update.ParameterFiles, "parameter-file", 'p', "path to parameter file")
	flags.FlagLong(&update.Yes, "yes", 'y', "do not prompt for stack update confirmation")
	flags.FlagLong(&update.StackName, "stack-name", 'n', "override inferrred stack name")
	flags.Parse(args)
	rest := flags.Args()

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
	_, err = cfn.CreateChangeSet(
		update.Sess(),
		update.StackName,
		string(template),
		parameters)
	if err != nil {
		return errors.Wrap(err, "failed to create change set")
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
