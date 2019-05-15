package main

import (
	"github.com/pkg/errors"
	"github.com/tetratom/cfn-tool/internal"
	"github.com/tetratom/cfn-tool/manifest"
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

	decision := manifest.Decision{}
	deployer, err := internal.NewDeployer(&prog.AWS, &decision)
	if err != nil {
		return errors.Wrap(err, "new deployer")
	}

	_, err = deployer.Whoami(w)
	return err
}
