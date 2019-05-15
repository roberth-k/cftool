package main

import (
	"github.com/pborman/getopt/v2"
	"github.com/pkg/errors"
	"github.com/tetratom/cftool/internal"
	"github.com/tetratom/cftool/manifest"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Update struct {
	Prog           *Program
	Parameters     []string
	ParameterFiles []string
	Yes            bool
	StackName      string
	TemplateFile   string
	ShowDiff       bool
}

func (update *Update) ParseFlags(args []string) error {
	flags := getopt.New()
	flags.FlagLong(&update.Parameters, "parameter", 'P', "explicit parameters")
	flags.FlagLong(&update.ParameterFiles, "parameter-file", 'p', "path to parameter file")
	flags.FlagLong(&update.Yes, "yes", 'y', "do not prompt for update confirmation (if a stack already exists)")
	flags.FlagLong(&update.StackName, "stack-name", 'n', "override inferrred stack name")
	flags.FlagLong(&update.TemplateFile, "template-file", 't', "template file")
	showDiff := flags.BoolLong("diff", 'd', "show template diff when updating a stack")
	flags.Parse(args)

	if showDiff != nil && *showDiff {
		update.ShowDiff = true
	}

	return nil
}

func (prog *Program) Update(args []string) error {
	update := Update{Prog: prog}
	err := update.ParseFlags(args)
	if err != nil {
		return err
	}

	err = prog.Whoami([]string{})
	if err != nil {
		return errors.Wrap(err, "get whoami")
	}

	stackName := update.deriveStackName()
	if stackName == "" {
		return errors.New("stack name is required")
	}

	parameters, err := update.parseAllParameters()
	if err != nil {
		return errors.Wrap(err, "parse parameters")
	}

	templateBody, err := ioutil.ReadFile(update.TemplateFile)
	if err != nil {
		return errors.Wrapf(err, "read %s", update.TemplateFile)
	}

	decision := manifest.Decision{
		AccountId:    "",
		Region:       "",
		TemplateBody: string(templateBody),
		Parameters:   parameters,
		StackName:    stackName,
		Protected:    !update.Yes,
	}

	deployer, err := internal.NewDeployer(&prog.AWS, &decision)
	if err != nil {
		return errors.Wrap(err, "new deployer")
	}

	deployer.ShowDiff = update.ShowDiff

	err = deployer.Deploy(w)
	if err != nil {
		return errors.Wrap(err, "deploy stack")
	}

	return nil
}

func (update *Update) deriveStackName() string {
	if update.StackName != "" {
		return update.StackName
	}

	getNameWithoutExtension := func(name string) string {
		basename := filepath.Base(name)
		noext := strings.Split(basename, ".")
		return noext[0]
	}

	for _, path := range update.ParameterFiles {
		name := getNameWithoutExtension(path)
		if name != "" {
			return name
		}
	}

	return getNameWithoutExtension(update.TemplateFile)
}

func (update *Update) parseAllParameters() (map[string]string, error) {
	files := update.ParameterFiles
	params := update.Parameters
	result := make(map[string]string)

	for _, path := range files {
		update.Prog.Verbosef("reading parameters from %s...", path)

		paramsFromFile, err := parseParameterFile(path)

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
		for _, param := range params {
			k, v := parseParameterString(param)
			result[k] = v
		}
	}

	return result, nil
}
