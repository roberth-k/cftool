package main

import (
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/tetratom/cfn-tool/cli/pprint"
	"log"
	"os"
	"strings"
)

// Whoami prints information about the caller's AWS principal.
func (prog *Program) Whoami(args []string) error {
	if len(args) > 1 {
		log.Printf("expected no additional arguments in: %s\n", strings.Join(args, " "))
		os.Exit(1)
	}

	stsClient := sts.New(prog.AWS.Session())
	output, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	if err != nil {
		return err
	}

	var region string
	regionPtr := prog.AWS.Session().Config.Region
	if regionPtr != nil {
		region = *regionPtr
	}

	pprint.Field("  Account", *output.Account)
	pprint.Field("     Role", *output.Arn)

	if region != "" || prog.Verbose {
		pprint.Field("   Region", region)
	}

	return nil
}
