package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cfn "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pborman/getopt/v2"
	"os"
)

type AWSOptions struct {
	Profile  string `short:"p" long:"profile" optional:"yes" description:"AWS credential profile"`
	Region   string `short:"r" long:"region" optional:"yes" description:"AWS region"`
	Endpoint string `short:"e" long:"endpoint" optional:"yes" description:"AWS API endpoint"`
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
	Verbose bool       `short:"v" long:"verbose" description:"Verbose output"`
}

func (prog *Program) MainArgs(argv []string) []string {
	set := getopt.New()

	set.Parse(argv)
	return set.Args()
}

func main() {
	prog := Program{}

	getopt.FlagLong(&prog.AWS.Region, "region", 'r', "AWS region")
	getopt.FlagLong(&prog.AWS.Profile, "profile", 'p', "AWS credential profile")
	getopt.FlagLong(&prog.AWS.Endpoint, "endpoint", 'e', "AWS API endpoint")
	getopt.FlagLong(&prog.Verbose, "verbose", 'v', "enable verbose output")
	getopt.ParseV2()
	rest := getopt.Args()

	if len(rest) < 1 {
		fmt.Printf("expected a subcommand: whoami, update.\n")
		os.Exit(1)
	}

	var err error

	switch rest[0] {
	case "whoami":
		err = prog.Whoami(rest[1:])

	default:
		fmt.Printf("unrecognized command: %s\n", rest[0])
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
