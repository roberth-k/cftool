package main

import (
	"fmt"
	"github.com/tetratom/cfn-tool/cli/pprint"
	"io/ioutil"
	"os"
)

type UpdateCommand struct {
	Parameters     []string                `short:"P" long:"parameter" optional:"yes"`
	ParameterFiles []string                `short:"p" long:"parameter-file" optional:"yes"`
	Yes            bool                    `short:"y" long:"yes"`
	StackName      string                  `short:"n" long:"stack-name" optional:"yes"`
	Positional     UpdateCommandPositional `positional-args:"yes"`
}

type UpdateCommandPositional struct {
	TemplateFile string `required:"yes"`
	Rest         []string
}

func (update *UpdateCommand) Execute(args []string) error {
	if len(update.Positional.Rest) != 0 {
		pprint.Errorf("Expected exactly one template file.")
		os.Exit(1)
	}

	_, err := parseAllParameters(
		updateCommand.ParameterFiles,
		updateCommand.Parameters)

	_ = PrintWhoami()

	pprint.Field("StackName", update.StackName)
	pprint.Field("TemplateFile", update.Positional.TemplateFile)

	template, err := ioutil.ReadFile(update.Positional.TemplateFile)

	if err != nil {
		return err
	}

	fmt.Printf("The template is: %s", string(template))

	return nil
}

var updateCommand UpdateCommand

func init() {
	_, _ = parser.AddCommand(
		"update",
		"Update a CloudFormation stack",
		"Update a CloudFormation stack.",
		&updateCommand)
}

func parseAllParameters(files []string, params []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, path := range files {
		if options.Verbose {
			pprint.Verbosef("Reading parameters from %s...", path)
		}

		paramsFromFile, err := ParseParameterFile(path)

		if err != nil {
			return nil, err
		}

		for k, v := range paramsFromFile {
			if _, ok := result[k]; ok && options.Verbose {
				pprint.Verbosef("Override parameter %s.", k)
			}

			result[k] = v
		}
	}

	if len(updateCommand.Parameters) > 0 {
		pprint.Verbosef("Applying command-line parameter overrides...")

		for _, paramSpec := range params {
			param := ParseParameterFromCommandLine(paramSpec)
			result[*param.ParameterKey] = *param.ParameterValue
		}
	}

	return result, nil
}
