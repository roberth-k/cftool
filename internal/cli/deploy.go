package cli

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/tetratom/cftool/internal"
	"github.com/tetratom/cftool/pkg/cftool"
	manifest2 "github.com/tetratom/cftool/pkg/manifest"
	"github.com/tetratom/cftool/pkg/pprint"
	"os"
	"path/filepath"
)

func Deploy(c context.Context, globalOpts GlobalOptions, deployOpts DeployOptions) (err error) {
	api, err := globalOpts.AWS.CloudFormationClient()
	if err != nil {
		return
	}

	stsapi, err := globalOpts.AWS.STSClient()
	if err != nil {
		return err
	}

	manifestPath := deployOpts.ManifestFile
	if manifestPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		manifestPath, err = findManifest(cwd)
		if err != nil {
			return err
		}
	}

	pprint.Field(color.Output, "Manifest", manifestPath)

	manifest, err := manifest2.ReadFromFile(manifestPath)
	if err != nil {
		return
	}

	if err = os.Chdir(filepath.Dir(manifestPath)); err != nil {
		return
	}

	var deployments []*cftool.Deployment

	if deployment, ok, err := manifest.FindDeployment(deployOpts.Tenant, deployOpts.Stack); err != nil {
		return err
	} else if ok {
		deployments = append(deployments, deployment)
	}

	for i, deployment := range deployments {
		if i > 0 {
			fmt.Fprint(color.Output, "\n")
		}

		deployer := internal.NewDeployer(api, deployment)
		deployer.ShowDiff = deployOpts.ShowDiff

		id, err := deployer.Whoami(color.Output, stsapi)
		if err != nil {
			return err
		}

		if deployment.AccountId != "" && deployment.AccountId != *id.Account {
			fmt.Fprintf(color.Output, "\nTenant account mismatch. Has the correct profile been selected?\n")
			os.Exit(1)
		}

		if !deployment.Protected && !deployOpts.Yes {
			deployment.Protected = true
		}

		if err = deployer.Deploy(c, color.Output); err != nil {
			return errors.Wrapf(err, "deploy stack: %s", deployment.StackName)
		}
	}

	return nil
}

func findManifest(startdir string) (result string, err error) {
	manifestName := ".cftool.yml"

	lastpath := ""
	reldir := ""
	for {
		newpath := filepath.Join(startdir, reldir, manifestName)

		if newpath == lastpath {
			// went all the way up to the root directory
			break
		}

		lastpath = newpath

		ok, err := fileExists(newpath)
		if err != nil {
			return "", err
		}

		if ok {
			return newpath, nil
		}

		reldir = filepath.Join(reldir, "..")
	}

	return "", errors.Errorf("manifest %s not found in any enclosing directory", manifestName)
}

func fileExists(path string) (ok bool, err error) {
	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
