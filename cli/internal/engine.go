package internal

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tetratom/cfn-tool/cli/pprint"
	"github.com/tetratom/cfn-tool/manifest"
	"io"
	"strings"
	"time"
)

var ErrAbortedByUser = errors.New("aborted by user")

type StackStatus string

func (status StackStatus) IsComplete() bool {
	return strings.HasSuffix(string(status), "_COMPLETE")
}

func (status StackStatus) IsFailed() bool {
	return strings.HasSuffix(string(status), "_FAILED")
}

func (status StackStatus) IsTerminal() bool {
	return status.IsComplete() || status.IsFailed()
}

// Engine visually executes CloudFormation deployments.
type Engine struct {
	sess   *session.Session
	client *cf.CloudFormation
}

func NewEngine(sess *session.Session) *Engine {
	return &Engine{sess, cf.New(sess)}
}

func (eng *Engine) Deploy(w io.Writer, d *manifest.Decision) error {
	pprint.Field(w, "StackName", d.StackName)

	exists, err := eng.stackExists(d.StackName)
	if err != nil {
		return errors.Wrapf(err, "describe stack %s", d.StackName)
	}

	if !exists {
		if !pprint.Promptf(w, "\nStack %s does not exist. Create?", d.StackName) {
			return ErrAbortedByUser
		}
	}

	nochange := false
	chset, err := eng.createChangeSet(d, !exists)
	if err != nil {
		if strings.Contains(err.Error(), "The submitted information didn't contain changes") {
			nochange = true
		} else {
			return errors.Wrap(err, "create change set")
		}
	}

	if nochange {
		fmt.Fprintf(w, "\nNo change.\n")
	} else {
		pprint.ChangeSet(w, chset)

		if d.Protected && !pprint.Promptf(w, "\nExecute change set?") {
			return ErrAbortedByUser
		}

		if chset == nil {
			return errors.New("expected non-nil chset")
		}

		since := time.Now()

		_, err = eng.client.ExecuteChangeSet(
			&cf.ExecuteChangeSetInput{
				StackName:     chset.StackName,
				ChangeSetName: chset.ChangeSetName,
			})
		if err != nil {
			return errors.Wrap(err, "execute change set")
		}

		stack, err := eng.monitorStackUpdate(w, d.StackName, since)
		if err != nil {
			return errors.Wrap(err, "monitor stack update")
		}

		status := StackStatus(*stack.StackStatus)
		if status.IsFailed() {
			return errors.New("stack update failed")
		}
	}

	outputs, err := eng.getStackOutputs(d.StackName)
	if err != nil {
		return errors.Wrap(err, "get stack outputs")
	}

	for i, output := range outputs {
		if i == 0 {
			fmt.Fprintf(w, "\n")
		}

		pprint.StackOutput(w, output)
	}

	return nil
}

func (eng *Engine) describeStack(name string) (*cf.Stack, error) {
	stacks, err := eng.client.DescribeStacks(
		&cf.DescribeStacksInput{StackName: aws.String(name)})

	if err != nil {
		return nil, errors.Wrapf(err, "describe stack %s", name)
	}

	if len(stacks.Stacks) != 1 {
		return nil, errors.Wrapf(err, "stack %s not found", name)
	}

	return stacks.Stacks[0], nil
}

func (eng *Engine) stackExists(name string) (bool, error) {
	_, err := eng.describeStack(name)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return false, nil
		}

		return false, err
	}

	return true, err
}

func (eng *Engine) createChangeSet(
	d *manifest.Decision,
	create bool,
) (*cf.DescribeChangeSetOutput, error) {

	changeSetType := cf.ChangeSetTypeUpdate
	if create {
		changeSetType = cf.ChangeSetTypeCreate
	}

	input := cf.CreateChangeSetInput{
		StackName:     aws.String(d.StackName),
		ChangeSetName: aws.String("StackUpdate-" + uuid.New().String()),
		Parameters:    make([]*cf.Parameter, len(d.Parameters)),
		TemplateBody:  aws.String(d.TemplateBody),
		ChangeSetType: aws.String(changeSetType),
		Capabilities: []*string{
			aws.String("CAPABILITY_IAM"),
			aws.String("CAPABILITY_NAMED_IAM"),
		},
	}

	index := 0
	for key, value := range d.Parameters {
		input.Parameters[index] = &cf.Parameter{
			ParameterKey:   aws.String(key),
			ParameterValue: aws.String(value),
		}

		index += 1
	}

	_, err := eng.client.CreateChangeSet(&input)
	if err != nil {
		return nil, err
	}

	var chset *cf.DescribeChangeSetOutput

	for done := false; !done; {
		// It's probably not going to be ready immediately anyway, so let's wait
		// at the start of the loop.
		time.Sleep(2 * time.Second)

		chset, err = eng.client.DescribeChangeSet(
			&cf.DescribeChangeSetInput{
				StackName:     input.StackName,
				ChangeSetName: input.ChangeSetName,
			})
		if err != nil {
			return nil, errors.Wrap(err, "describe change set")
		}

		switch *chset.Status {
		case cf.ChangeSetStatusCreateComplete:
			done = true

		case cf.ChangeSetStatusFailed:
			return nil, errors.Errorf(
				"failed to create change set: %s", *chset.StatusReason)

		case cf.ChangeSetStatusDeleteComplete:
			return nil, errors.New("change set removed unexpectedly")
		}
	}

	return chset, nil
}

func (eng *Engine) getStackEvents(
	stackName string,
	since time.Time,
	until time.Time,
) ([]*cf.StackEvent, error) {
	out, err := eng.client.DescribeStackEvents(
		&cf.DescribeStackEventsInput{
			StackName: aws.String(stackName),
		})
	if err != nil {
		return nil, errors.Wrap(err, "describe stack events")
	}

	result := make([]*cf.StackEvent, 0, len(out.StackEvents))
	for _, event := range out.StackEvents {
		if (event.Timestamp.After(since) || event.Timestamp.Equal(since)) &&
			event.Timestamp.Before(until) {

			result = append(result, event)
		}
	}

	return result, nil
}

func (eng *Engine) getStackOutputs(stackName string) ([]*cf.Output, error) {
	stack, err := eng.client.DescribeStacks(
		&cf.DescribeStacksInput{
			StackName: aws.String(stackName),
		})
	if err != nil {
		return nil, errors.Wrap(err, "describe stack")
	}

	return stack.Stacks[0].Outputs, nil
}

func (eng *Engine) monitorStackUpdate(
	w io.Writer,
	stackName string,
	startTime time.Time,
) (stack *cf.Stack, err error) {
	lastStatus := StackStatus("UNKNOWN")
	since := startTime

	for i := 0; ; i++ {
		stack, err = eng.describeStack(stackName)
		if err != nil {
			return nil, err
		}

		if stack == nil {
			return nil, errors.New("unexpected nil stack")
		}

		status := StackStatus(*stack.StackStatus)

		if status != lastStatus {
			fmt.Fprintf(w, "\n")
			t := time.Now()
			events, err := eng.getStackEvents(stackName, since, t)
			since = t
			if err != nil {
				return nil, errors.Wrap(err, "get stack events")
			}

			for _, event := range events {
				if strings.HasSuffix(*event.ResourceStatus, "_FAILED") {
					pprint.StackEvent(w, event)
				}
			}

			lastStatus, i = status, 0
			fmt.Fprintf(w, "%s", status)

			if !status.IsTerminal() {
				fmt.Fprintf(w, "...")
			}
		}

		if status.IsTerminal() {
			fmt.Fprintf(w, "\n")
			break
		}

		sleepTime := 5 * time.Second

		if i < 5 {
			// Rapid updates for the first 10 seconds.
			sleepTime = 2 * time.Second
		}

		time.Sleep(sleepTime)
		fmt.Fprintf(w, ".")
	}

	return stack, err
}
