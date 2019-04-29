package cfn

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type StackStatus string

//const (
//	StackStatusCreateInProgress                        StackStatus = "CREATE_IN_PROGRESS"
//	StackStatusCreateFailed                            StackStatus = "CREATE_FAILED"
//	StackStatusCreateComplete                          StackStatus = "CREATE_COMPLETE"
//	StackStatusRollbackInProgress                      StackStatus = "ROLLBACK_IN_PROGRESS"
//	StackStatusRollbackFailed                          StackStatus = "ROLLBACK_FAILED"
//	StackStatusRollbackComplete                        StackStatus = "ROLLBACK_COMPLETE"
//	StackStatusDeleteInProgress                        StackStatus = "DELETE_IN_PROGRESS"
//	StackStatusDeleteFailed                            StackStatus = "DELETE_FAILED"
//	StackStatusDeleteComplete                          StackStatus = "DELETE_COMPLETE"
//	StackStatusUpdateInProgress                        StackStatus = "UPDATE_IN_PROGRESS"
//	StackStatusUpdateCompleteCleanupInProgress         StackStatus = "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS"
//	StackStatusUpdateComplete                          StackStatus = "UPDATE_COMPLETE"
//	StackStatusUpdateRollbackInProgress                StackStatus = "UPDATE_ROLLBACK_IN_PROGRESS"
//	StackStatusUpdateRollbackFailed                    StackStatus = "UPDATE_ROLLBACK_FAILED"
//	StackStatusUpdateRollbackCompleteCleanupInProgress StackStatus = "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS"
//	StackStatusUpdateRollbackComplete                  StackStatus = "UPDATE_ROLLBACK_COMPLETE"
//	StackStatusReviewInProgress                        StackStatus = "REVIEW_IN_PROGRESS"
//)

func (status StackStatus) IsComplete() bool {
	return strings.HasSuffix(string(status), "_COMPLETE")
}

func (status StackStatus) IsFailed() bool {
	return strings.HasSuffix(string(status), "_FAILED")
}

func (status StackStatus) IsTerminal() bool {
	return status.IsComplete() || status.IsFailed()
}

type StackUpdate struct {
	Name      string
	StackName string
	Id        string
	StackId   string
	client    *cf.CloudFormation
}

func CreateChangeSet(
	sess *session.Session,
	stackName string,
	templateBody string,
	parameters map[string]string,
	create bool,
) (*StackUpdate, error) {

	client := cf.New(sess)

	changeSetType := cf.ChangeSetTypeUpdate
	if create {
		changeSetType = cf.ChangeSetTypeCreate
	}

	input := cf.CreateChangeSetInput{
		StackName:     aws.String(stackName),
		ChangeSetName: aws.String("StackUpdate-" + uuid.New().String()),
		Parameters:    make([]*cf.Parameter, len(parameters)),
		TemplateBody:  aws.String(templateBody),
		ChangeSetType: aws.String(changeSetType),
		Capabilities: []*string{
			aws.String("CAPABILITY_IAM"),
			aws.String("CAPABILITY_NAMED_IAM"),
		},
	}

	index := 0
	for key, value := range parameters {
		input.Parameters[index] = &cf.Parameter{
			ParameterKey:   aws.String(key),
			ParameterValue: aws.String(value),
		}

		index += 1
	}

	out, err := client.CreateChangeSet(&input)
	if err != nil {
		return nil, err
	}

	update := StackUpdate{
		Name:      *input.ChangeSetName,
		StackName: *input.StackName,
		Id:        *out.Id,
		StackId:   *out.StackId,
		client:    client,
	}

	for done := false; !done; {
		status, err := update.DescribeChangeSet()
		if err != nil {
			return nil, errors.Wrap(err, "describe change set")
		}

		switch *status.Status {
		case cf.ChangeSetStatusCreateComplete:
			done = true

		case cf.ChangeSetStatusFailed:
			return nil, errors.Errorf(
				"failed to create change set: %s", *status.StatusReason)

		case cf.ChangeSetStatusDeleteComplete:
			return nil, errors.New("change set removed unexpectedly")

		default:
			time.Sleep(2 * time.Second)
		}
	}

	return &update, nil
}

func (update *StackUpdate) DescribeChangeSet() (*cf.DescribeChangeSetOutput, error) {
	out, err := update.client.DescribeChangeSet(
		&cf.DescribeChangeSetInput{
			StackName:     aws.String(update.StackName),
			ChangeSetName: aws.String(update.Name)})
	if err != nil {
		return nil, errors.Wrap(err, "describe change set")
	}

	return out, nil
}

func (update *StackUpdate) DescribeStack() (*cf.Stack, error) {
	out, err := update.client.DescribeStacks(
		&cf.DescribeStacksInput{
			StackName: aws.String(update.StackName)})
	if err != nil {
		return nil, errors.Wrap(err, "describe stack")
	}

	if len(out.Stacks) != 1 {
		return nil, errors.New("expected to find exactly one stack")
	}

	return out.Stacks[0], nil
}

func (update *StackUpdate) GetStatus() (StackStatus, error) {
	out, err := update.client.DescribeStacks(&cf.DescribeStacksInput{
		StackName: aws.String(update.StackId),
	})

	if err != nil {
		return "", errors.Wrapf(err, "describe stack %s", update.StackId)
	}

	stack := out.Stacks[0]
	status := StackStatus(*stack.StackStatus)
	return status, nil
}

func (update *StackUpdate) Execute() error {
	_, err := update.client.ExecuteChangeSet(&cf.ExecuteChangeSetInput{
		StackName:     aws.String(update.StackId),
		ChangeSetName: aws.String(update.Name),
	})

	if err != nil {
		return errors.Wrap(err, "execute change set")
	}

	return nil
}

func StackExists(sess *session.Session, stackName string) (bool, error) {
	client := cf.New(sess)
	stack, err := client.DescribeStacks(&cf.DescribeStacksInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return false, nil
		}

		return false, err
	}

	if len(stack.Stacks) > 0 {
		if *stack.Stacks[0].StackStatus == cf.StackStatusReviewInProgress {
			return false, nil
		}
	}

	return true, nil
}

func (update *StackUpdate) GetEvents(
	since time.Time,
	until time.Time,
) ([]*cf.StackEvent, error) {
	out, err := update.client.DescribeStackEvents(
		&cf.DescribeStackEventsInput{
			StackName: aws.String(update.StackName),
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

func (update *StackUpdate) GetOutputs() ([]*cf.Output, error) {
	stack, err := update.DescribeStack()
	if err != nil {
		return nil, errors.Wrap(err, "describe stack")
	}

	return stack.Outputs, nil
}
