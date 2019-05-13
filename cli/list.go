package main

import (
	"fmt"
	"github.com/pborman/getopt/v2"
	"github.com/pkg/errors"
	"github.com/tetratom/cfn-tool/cli/pprint"
	"github.com/tetratom/cfn-tool/manifest"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"
)

type List struct {
	ManifestFile string
}

func (list *List) ParseArgs(args []string) error {
	flags := getopt.New()
	flags.FlagLong(&list.ManifestFile, "manifest", 'f', "manifest path")
	flags.Parse(args)
	if len(flags.Args()) != 0 {
		return errors.Errorf("unexpected positional arguments: %+v", flags.Args())
	}

	if list.ManifestFile == "" {
		manifest, err := findManifest()
		if err != nil {
			return err
		}
		list.ManifestFile = manifest
	}
	return nil
}

type DeploymentsByStackName []*manifest.Decision

func (d DeploymentsByStackName) Len() int {
	return len(d)
}

func (d DeploymentsByStackName) Less(i, j int) bool {
	return d[i].Stack.Name < d[j].Stack.Name
}

func (d DeploymentsByStackName) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (prog *Program) List(args []string) error {
	list := List{}
	if err := list.ParseArgs(args); err != nil {
		return err
	}

	pprint.Field(os.Stdout, "Manifest", list.ManifestFile)

	fp, err := os.Open(list.ManifestFile)
	if err != nil {
		return errors.Wrapf(err, "open %s", list.ManifestFile)
	}

	if err := os.Chdir(filepath.Dir(list.ManifestFile)); err != nil {
		return err
	}

	m, err := manifest.Parse(fp)
	fp.Close()
	if err != nil {
		return errors.Wrap(err, "parse manifest")
	}

	all, err := m.AllDeployments()
	if err != nil {
		return errors.Wrap(err, "list all deployments")
	}

	sort.Sort(DeploymentsByStackName(all))

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintf(w, "Stack\tTenant\n")
	fmt.Fprintf(w, "---\t---\n")

	for i, dep := range all {
		if i == 0 {
			fmt.Printf("\n")
		}

		fmt.Fprintf(w, "%s\t%s\n", dep.Stack.Name, dep.Tenant.Name)
	}

	_ = w.Flush()

	return nil
}
