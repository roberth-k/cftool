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
