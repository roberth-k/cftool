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

	id, err := prog.getWhoami()
	if err != nil {
		return err
	}

	pprint.Whoami(os.Stdout, prog.AWS.Session().Config.Region, id)
	return nil
}

func (prog *Program) getWhoami() (*sts.GetCallerIdentityOutput, error) {
	client := sts.New(prog.AWS.Session())
	identity, err := client.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}

	return identity, nil
}
