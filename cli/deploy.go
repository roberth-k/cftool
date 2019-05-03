package main

import (
	"fmt"
	"github.com/pborman/getopt/v2"
	"github.com/pkg/errors"
	"github.com/tetratom/cfn-tool/cli/internal"
	"github.com/tetratom/cfn-tool/cli/pprint"
	"github.com/tetratom/cfn-tool/manifest"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Deploy struct {
	Prog         *Program
	Yes          bool
	ManifestFile string
	Stack        string
	Tenant       string
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func findManifest(file string) (string, error) {
	file = strings.TrimSpace(file)

	if file == "" {
		file = ".cfn-tool.yml"

		for {
			ok, err := fileExists(file)
			if err != nil {
				fmt.Printf("%s\n", errors.Wrapf(err, "check %s", file))
				os.Exit(1)
			}

			if ok {
				break
			}

			file = path.Join("..", file)
			abspath, _ := filepath.Abs(file)
			if filepath.Dir(abspath) == "/" {
				fmt.Printf("unable to find .cfn-tool.yml\n")
				os.Exit(1)
			}
		}
	}

	file = strings.TrimSpace(file)

	if strings.HasPrefix(file, "~/") {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Errorf("unable to determine user home directory")
		}

		file = homedir + file[1:]
	}

	if ok, _ := fileExists(file); !ok {
		return "", errors.Errorf("%s does not exist\n", file)
	}

	return file, nil
}

func (d *Deploy) ParseArgs(args []string) []string {
	flags := getopt.New()
	flags.FlagLong(&d.Yes, "yes", 'y', "do not prompt for confirmation")
	flags.FlagLong(&d.ManifestFile, "manifest", 'f', "manifest path")
	flags.FlagLong(&d.Stack, "stack", 's', "stack to deploy")
	flags.FlagLong(&d.Tenant, "tenant", 't', "tenant to deploy for")
	flags.Parse(args)
	rest := flags.Args()

	if len(rest) != 0 {
		fmt.Printf("Did not expect positional parameters.\n")
		os.Exit(1)
	}

	manifest, err := findManifest(d.ManifestFile)
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	d.ManifestFile = manifest

	return rest
}

func (prog *Program) Deploy(args []string) error {
	d := Deploy{}
	d.ParseArgs(args)

	pprint.Field(os.Stdout, "Manifest", d.ManifestFile)

	fp, err := os.Open(d.ManifestFile)
	if err != nil {
		return errors.Wrapf(err, "open %s", d.ManifestFile)
	}

	m, err := manifest.Parse(fp)
	fp.Close()
	if err != nil {
		return errors.Wrap(err, "parse manifest")
	}

	err = os.Chdir(filepath.Dir(d.ManifestFile))
	if err != nil {
		return errors.Wrap(err, "chdir")
	}

	decisions, err := m.Process(manifest.ProcessInput{
		Stack:  d.Stack,
		Tenant: d.Tenant})
	if err != nil {
		return errors.Wrap(err, "process manifest")
	}

	for _, decision := range decisions {
		fmt.Printf("\n")

		deployer, err := internal.NewDeployer(&prog.AWS, decision)
		if err != nil {
			return errors.Wrap(err, "new deployer")
		}

		err = deployer.Whoami(os.Stdout)
		if err != nil {
			return errors.Wrap(err, "whoami")
		}

		if !decision.Protected && !d.Yes {
			decision.Protected = true
		}

		err = deployer.Deploy(os.Stdout)
		if err != nil {
			return errors.Wrap(err, "deploy stack")
		}
	}

	return nil
}
