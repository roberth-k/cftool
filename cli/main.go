package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jessevdk/go-flags"
	"os"
)

type AWSOptions struct {
	Profile  string `short:"p" long:"profile" optional:"yes" description:"AWS credential profile"`
	Region   string `short:"r" long:"region" optional:"yes" description:"AWS region"`
	Endpoint string `short:"e" long:"endpoint" optional:"yes" description:"AWS API endpoint"`
}

type Options struct {
	AWS     AWSOptions `group:"AWS options"`
	Verbose bool       `short:"v" long:"verbose" description:"Verbose output"`
}

var options Options
var parser = flags.NewParser(&options, flags.Default)
var awsSession *session.Session

func main() {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok {
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		} else {
			os.Exit(1)
		}
	}
}

func GetAWSSession() *session.Session {
	if awsSession == nil {
		sessionOptions := session.Options{
			Config: aws.Config{
				Endpoint: aws.String(options.AWS.Endpoint),
				Region:   aws.String(options.AWS.Region),
			},
			Profile: options.AWS.Profile,
		}

		awsSession = session.Must(session.NewSessionWithOptions(sessionOptions))
	}

	return awsSession
}
