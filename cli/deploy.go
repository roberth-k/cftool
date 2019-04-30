package main

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pborman/getopt/v2"
	"github.com/pkg/errors"
	"github.com/tetratom/cfn-tool/manifest"
	"io/ioutil"
	"os"
)

type Deploy struct {
	Prog         *Program
	Yes          bool
	ManifestFile string
}

func (d *Deploy) ParseArgs(args []string) []string {
	flags := getopt.New()
	flags.FlagLong(&d.Yes, "yes", 'y', "do not prompt for confirmation")
	flags.FlagLong(&d.ManifestFile, "tool-file", 'f', "tool file path")
	flags.Parse(args)
	rest := flags.Args()

	if len(rest) != 0 {
		fmt.Printf("Did not expect positional parameters.\n")
		os.Exit(1)
	}

	return rest
}

func (prog *Program) Deploy(args []string) error {
	d := Deploy{}
	d.ParseArgs(args)

	m, err := d.parseManifest()
	if err != nil {
		return errors.Wrap(err, "parse manifest")
	}

	if m.Version != manifest.Version1_0 {
		return errors.Errorf("expected manifest version %s", manifest.Version1_0)
	}

	fmt.Printf("%+v\n", *m)
	return nil
}

func (d *Deploy) parseManifest() (*manifest.Manifest, error) {
	data, err := ioutil.ReadFile(d.ManifestFile)
	if err != nil {
		return nil, errors.Wrap(err, "read manifest file")
	}

	var m manifest.Manifest
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal manifest")
	}

	return &m, nil
}
