package main

import (
	"fmt"
	"github.com/pborman/getopt/v2"
	"github.com/pkg/errors"
	"github.com/tetratom/cfn-tool/internal"
	"github.com/tetratom/cfn-tool/manifest"
	"github.com/tetratom/cfn-tool/pprint"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Deploy struct {
	Prog         *Program
	Yes          bool
	ManifestFile string
	Stack        string
	Tenant       string
	ShowDiff     bool
}

var fs struct {
	Getwd       func() (cwd string, err error)
	UserHomeDir func() (dir string, err error)
	FileExists  func(path string) (ok bool, err error)
	ExpandUser  func(path string) (out string, err error)
	VolumeName  func(path string) (out string)
}

func init() {
	fs.Getwd = func() (cwd string, err error) {
		return os.Getwd()
	}

	fs.UserHomeDir = func() (dir string, err error) {
		return os.UserHomeDir()
	}

	fs.FileExists = func(path string) (ok bool, err error) {
		_, err = os.Stat(path)

		if os.IsNotExist(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return true, nil
	}

	fs.ExpandUser = func(path string) (out string, err error) {
		path = strings.TrimSpace(path)
		if strings.HasPrefix(path, "~/") {
			homedir, err := fs.UserHomeDir()
			if err != nil {
				return "", errors.Errorf("unable to determine user home directory")
			}

			path = homedir + path[1:]
		}
		return path, nil
	}

	fs.VolumeName = func(path string) (out string) {
		return filepath.VolumeName(path)
	}
}

func findManifest() (string, error) {
	dir, filename := "", ".cfn-tool.yml"
	cwd, err := fs.Getwd()
	rootdir := fs.VolumeName(cwd)

	if rootdir == "" {
		rootdir = "/"
	}

	if err != nil {
		log.Panic(errors.Wrap(err, "getwd"))
	}

	for {
		dirabs := filepath.Join(cwd, dir)
		fpath := filepath.Join(dirabs, filename)

		ok, err := fs.FileExists(fpath)
		if err != nil {
			log.Panic(errors.Wrap(err, "get file exists"))
		}

		if ok {
			return filepath.Join(dir, filename), nil
		}

		if dirabs == rootdir {
			return "", errors.Errorf(
				"manifest %s not found in any enclosing directory",
				filename)
		}

		dir = filepath.Join(dir, "..")
	}
}

func (d *Deploy) ParseArgs(args []string) []string {
	flags := getopt.New()
	flags.FlagLong(&d.Yes, "yes", 'y', "do not prompt for confirmation")
	flags.FlagLong(&d.ManifestFile, "manifest", 'f', "manifest path")
	flags.FlagLong(&d.Stack, "stack", 's', "stack to deploy")
	flags.FlagLong(&d.Tenant, "tenant", 't', "tenant to deploy for")
	showDiff := flags.BoolLong("diff", 'd', "show template diff when updating a stack")
	flags.Parse(args)
	rest := flags.Args()

	if len(rest) != 0 {
		fmt.Printf("Did not expect positional parameters.\n")
		os.Exit(1)
	}

	if showDiff != nil && *showDiff {
		d.ShowDiff = true
	}

	if d.ManifestFile == "" {
		manifest, err := findManifest()
		if err != nil {
			fmt.Printf(err.Error())
			os.Exit(1)
		}
		d.ManifestFile = manifest
	}

	return rest
}

func (prog *Program) Deploy(args []string) error {
	d := Deploy{}
	d.ParseArgs(args)

	pprint.Field(w, "Manifest", d.ManifestFile)

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
		Tenant: d.Tenant,
	})

	if err != nil {
		return errors.Wrap(err, "process manifest")
	}

	for i, decision := range decisions {
		if i > 0 {
			fmt.Printf("\n")
		}

		deployer, err := internal.NewDeployer(&prog.AWS, decision)
		if err != nil {
			return errors.Wrap(err, "new deployer")
		}

		deployer.ShowDiff = d.ShowDiff

		id, err := deployer.Whoami(w)
		if err != nil {
			return errors.Wrap(err, "whoami")
		}

		if decision.AccountId != "" && decision.AccountId != *id.Account {
			fmt.Fprintf(w, "\nTenant account mismatch. Has the correct profile been selected?\n")
			os.Exit(1)
		}

		if !decision.Protected && !d.Yes {
			decision.Protected = true
		}

		err = deployer.Deploy(w)
		if err != nil {
			return errors.Wrap(err, "deploy stack")
		}
	}

	return nil
}
