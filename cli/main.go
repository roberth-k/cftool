package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cfn "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pborman/getopt/v2"
	"github.com/pkg/errors"
	"github.com/tetratom/cfn-tool/cli/internal"
	"github.com/tetratom/cfn-tool/cli/pprint"
	"os"
)

type AWSOptions struct {
	Profile  string
	Region   string
	Endpoint string
	sess     *session.Session
	client   *cfn.CloudFormation
}

func (opts *AWSOptions) CfnClient() *cfn.CloudFormation {
	if opts.client == nil {
		opts.client = cfn.New(opts.Session())
	}

	return opts.client
}

func (opts *AWSOptions) Session() *session.Session {
	if opts.sess == nil {
		sessOpts := session.Options{Config: aws.Config{}}

		_ = os.Setenv("AWS_SDK_LOAD_CONFIG", "1")

		if opts.Profile != "" {
			sessOpts.Profile = opts.Profile
		}

		if opts.Region != "" {
			sessOpts.Config.Region = aws.String(opts.Region)
		}

		if opts.Endpoint != "" {
			sessOpts.Config.Endpoint = aws.String(opts.Region)
		}

		opts.sess = session.Must(session.NewSessionWithOptions(sessOpts))
	}

	return opts.sess
}

type Program struct {
	AWS     AWSOptions `group:"AWS options"`
	Verbose bool
	Color   bool
}

func (prog *Program) ParseFlags(argv []string) []string {
	set := getopt.New()
	set.FlagLong(&prog.AWS.Region, "region", 'r', "AWS region")
	set.FlagLong(&prog.AWS.Profile, "profile", 'p', "AWS credential profile")
	set.FlagLong(&prog.AWS.Endpoint, "endpoint", 'e', "AWS API endpoint")
	set.FlagLong(&prog.Verbose, "verbose", 'v', "enable verbose output")
	color := set.EnumLong(
		"color", 'c', []string{"on", "off"}, "on",
		"'on' or 'off'. pass 'off' to disable colors.")
	set.Parse(argv)

	if color == nil || *color == "on" {
		prog.Color = true
	}

	return set.Args()
}

func main() {
	prog := Program{}
	rest := prog.ParseFlags(os.Args)

	if len(rest) < 1 {
		fmt.Printf("Expected a subcommand: deploy, update, or whoami.\n")
		os.Exit(1)
	}

	if !prog.Color {
		pprint.DisableColor()
	}

	var err error

	switch rest[0] {
	case "deploy":
		err = prog.Deploy(rest)

	case "update":
		err = prog.Update(rest)

	case "whoami":
		err = prog.Whoami(rest)

	default:
		fmt.Printf("unrecognized command: %s\n", rest[0])
		os.Exit(1)
	}

	if err != nil {
		if errors.Cause(err) == internal.ErrAbortedByUser {
			fmt.Fprintf(os.Stderr, "Aborted by user.\n")
			os.Exit(1)
		}

		fmt.Printf("error: %s\n", err) // TODO: %+v
		os.Exit(1)
	}

	os.Exit(0)
}

func (p *Program) Verbosef(msg string, args ...interface{}) {
	if p.Verbose {
		pprint.Verbosef(os.Stdout, msg, args...)
	}
}
