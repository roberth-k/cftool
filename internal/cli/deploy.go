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

	manifestPath := deployOpts.ManifestFile
	if manifestPath == "" {
		manifestPath, err = findManifest()
		if err != nil {
			return
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

		if i == 0 {
			if _, err := deployer.Whoami(color.Output); err != nil {
				return err
			}
		}

		// todo: check account and region match

		if !deployment.Protected && !deployOpts.Yes {
			deployment.Protected = true
		}

		if err = deployer.Deploy(color.Output); err != nil {
			return errors.Wrapf(err, "deploy stack: %s", deployment.StackName)
		}
	}

	return nil
}

func findManifest() (string, error) {
	panic("not implemented")
}
