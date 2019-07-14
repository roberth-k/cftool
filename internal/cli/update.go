package cli

import (
	"context"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/tetratom/cftool/internal"
	"github.com/tetratom/cftool/pkg/cftool"
	"github.com/tetratom/cftool/pkg/manifest"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func Update(c context.Context, globalOpts GlobalOptions, updateOpts UpdateOptions) (err error) {
	api, err := globalOpts.AWS.CloudFormationClient()
	if err != nil {
		return
	}

	stackName, err := deriveStackName(updateOpts)
	if err != nil {
		return
	}

	parameters, err := parseParameters(updateOpts)
	if err != nil {
		return
	}

	templateBody, err := ioutil.ReadFile(updateOpts.TemplateFile)
	if err != nil {
		return errors.Wrapf(err, "read template: %s", updateOpts.TemplateFile)
	}

	deployment := cftool.Deployment{
		AccountId:    "",
		Region:       "",
		TemplateBody: templateBody,
		Parameters:   parameters,
		StackName:    string(stackName), // todo: type conversion
		Protected:    !updateOpts.Yes,
	}

	deployer := internal.NewDeployer(api, &deployment)
	deployer.ShowDiff = updateOpts.ShowDiff

	stsapi, err := globalOpts.AWS.STSClient()
	if err != nil {
		return err
	}

	if _, err := deployer.Whoami(color.Output, stsapi); err != nil {
		return err
	}

	if err = deployer.Deploy(c, color.Output); err != nil {
		return errors.Wrapf(err, "deploy stack: %s", stackName)
	}

	return nil
}

func deriveStackName(opts UpdateOptions) (cftool.StackName, error) {
	if opts.StackName != "" {
		return cftool.StackName(opts.StackName), nil
	}

	list := []struct {
		names []string
	}{
		{opts.ParameterFiles},
		{[]string{opts.TemplateFile}},
	}

	for _, element := range list {
		if len(element.names) > 0 {
			basename := filepath.Base(opts.ParameterFiles[0])
			return cftool.StackName(strings.Split(basename, ".")[0]), nil
		}
	}

	return "", errors.New("unable to derive stack name")
}

func parseParameters(update UpdateOptions) (cftool.Parameters, error) {
	files := update.ParameterFiles
	params := update.Parameters
	result := make(map[string]string)

	for _, path := range files {
		paramsFromFile, err := manifest.ReadParametersFromFile(path)

		if err != nil {
			return nil, err
		}

		for k, v := range paramsFromFile {
			result[k] = v
		}
	}

	if len(update.Parameters) > 0 {
		for _, param := range params {
			k, v := parseParameterString(param)
			result[k] = v
		}
	}

	return result, nil
}

func parseParameterString(str string) (string, string) {
	split := strings.SplitN(str, "=", 2)
	key := split[0]
	value := ""

	if len(split) > 1 {
		value = split[1]
	}

	return key, value
}
