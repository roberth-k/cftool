package cli

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/pborman/getopt/v2"
	"github.com/pkg/errors"
	"github.com/tetratom/cftool/internal"
	"os"
)

type GlobalOptions struct {
	AWS           AWSOptions
	Color         bool
	Version       bool
	remainingArgs []string
}

type AWSOptions struct {
	Profile  string
	Region   string
	Endpoint string
}

func (awsOpts *AWSOptions) CloudFormationClient() (cloudformationiface.CloudFormationAPI, error) {
	opts := session.Options{}
	opts.SharedConfigState = session.SharedConfigEnable
	opts.AssumeRoleTokenProvider = stscreds.StdinTokenProvider

	if awsOpts.Profile != "" {
		opts.Profile = awsOpts.Profile
	}

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return nil, errors.Wrap(err, "create aws session")
	}

	creds, err := internal.WrapCredentialsWithCache(opts.Profile, sess.Config.Credentials)
	if err != nil {
		return nil, errors.Wrap(err, "credential cache")
	}

	sess.Config.Credentials = creds

	return cloudformation.New(sess), nil
}

func ParseGlobalOptions(args []string) GlobalOptions {
	var options GlobalOptions

	flags := getopt.New()
	flags.FlagLong(&options.AWS.Region, "region", 'r', "AWS region")
	flags.FlagLong(&options.AWS.Profile, "profile", 'p', "AWS credential profile")
	flags.FlagLong(&options.AWS.Endpoint, "endpoint", 'e', "AWS API endpoint")

	color := flags.EnumLong(
		"color", 'c', []string{"on", "off"}, "on",
		"'on' or 'off'. pass 'off' to disable colors.")

	flags.FlagLong(&options.Version, "version", 'V', "show version and exit")

	flags.Parse(args)

	if color == nil || *color == "on" {
		options.Color = true
	}

	options.remainingArgs = flags.Args()

	return options
}

type DeployOptions struct {
	Yes          bool
	ManifestFile string
	Stack        string
	Tenant       string
	ShowDiff     bool
}

func ParseDeployOptions(args []string) DeployOptions {
	var options DeployOptions

	flags := getopt.New()
	flags.FlagLong(&options.Yes, "yes", 'y', "do not prompt for confirmation")
	flags.FlagLong(&options.ManifestFile, "manifest", 'f', "manifest path")
	flags.FlagLong(&options.Stack, "stack", 's', "stack to deploy")
	flags.FlagLong(&options.Tenant, "tenant", 't', "tenant to deploy for")
	showDiff := flags.BoolLong("diff", 'd', "show template diff when updating a stack")
	flags.Parse(args)
	rest := flags.Args()

	if len(rest) != 0 {
		fmt.Printf("Did not expect positional parameters.\n")
		os.Exit(1)
	}

	if showDiff != nil && *showDiff {
		options.ShowDiff = true
	}

	return options
}

type UpdateOptions struct {
	Parameters     []string
	ParameterFiles []string
	Yes            bool
	StackName      string
	TemplateFile   string
	ShowDiff       bool
}

func ParseUpdateOptions(args []string) UpdateOptions {
	var options UpdateOptions

	flags := getopt.New()
	flags.FlagLong(&options.Parameters, "parameter", 'P', "explicit parameters")
	flags.FlagLong(&options.ParameterFiles, "parameter-file", 'p', "path to parameter file")
	flags.FlagLong(&options.Yes, "yes", 'y', "do not prompt for update confirmation (if a stack already exists)")
	flags.FlagLong(&options.StackName, "stack-name", 'n', "override inferrred stack name")
	flags.FlagLong(&options.TemplateFile, "template-file", 't', "template file")
	showDiff := flags.BoolLong("diff", 'd', "show template diff when updating a stack")
	flags.Parse(args)

	if showDiff != nil && *showDiff {
		options.ShowDiff = true
	}

	return options
}
