package main

import (
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/tetratom/cfn-tool/cli/pprint"
)

type WhoamiCommand struct {
}

func (command *WhoamiCommand) Execute(args []string) error {
	return PrintWhoami()
}

var whoamiCommand WhoamiCommand

func init() {
	_, _ = parser.AddCommand(
		"whoami",
		"Show AWS caller identity",
		"Show AWS caller identity based on the current credentials.",
		&whoamiCommand)
}

// PrintWhoami prints information about the caller's AWS principal.
func PrintWhoami() error {
	stsClient := sts.New(GetAWSSession())
	output, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	if err != nil {
		return err
	}

	region := *GetAWSSession().Config.Region

	pprint.Field("Account", *output.Account)
	pprint.Field("   Role", *output.Arn)

	if region != "" || options.Verbose {
		pprint.Field(" Region", region)
	}

	return nil
}
