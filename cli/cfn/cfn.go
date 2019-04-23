package cfn

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/google/uuid"
	"strings"
)

type StackStatus string

const (
	StackStatusCreateInProgress                        StackStatus = "CREATE_IN_PROGRESS"
	StackStatusCreateFailed                            StackStatus = "CREATE_FAILED"
	StackStatusCreateComplete                          StackStatus = "CREATE_COMPLETE"
	StackStatusRollbackInProgress                      StackStatus = "ROLLBACK_IN_PROGRESS"
	StackStatusRollbackFailed                          StackStatus = "ROLLBACK_FAILED"
	StackStatusRollbackComplete                        StackStatus = "ROLLBACK_COMPLETE"
	StackStatusDeleteInProgress                        StackStatus = "DELETE_IN_PROGRESS"
	StackStatusDeleteFailed                            StackStatus = "DELETE_FAILED"
	StackStatusDeleteComplete                          StackStatus = "DELETE_COMPLETE"
	StackStatusUpdateInProgress                        StackStatus = "UPDATE_IN_PROGRESS"
	StackStatusUpdateCompleteCleanupInProgress         StackStatus = "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS"
	StackStatusUpdateComplete                          StackStatus = "UPDATE_COMPLETE"
	StackStatusUpdateRollbackInProgress                StackStatus = "UPDATE_ROLLBACK_IN_PROGRESS"
	StackStatusUpdateRollbackFailed                    StackStatus = "UPDATE_ROLLBACK_FAILED"
	StackStatusUpdateRollbackCompleteCleanupInProgress StackStatus = "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS"
	StackStatusUpdateRollbackComplete                  StackStatus = "UPDATE_ROLLBACK_COMPLETE"
	StackStatusReviewInProgress                        StackStatus = "REVIEW_IN_PROGRESS"
)

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
	Name    string
	Id      string
	StackId string
	client  *cloudformation.CloudFormation
}

func CreateChangeSet(
	sess *session.Session,
	stackName string,
	templateBody string,
	parameters map[string]string) (StackUpdate, error) {

	client := cloudformation.New(sess)

	input := cloudformation.CreateChangeSetInput{
		StackName:     aws.String(stackName),
		ChangeSetName: aws.String("StackUpdate-" + uuid.New().String()),
		Parameters:    make([]*cloudformation.Parameter, len(parameters)),
		TemplateBody:  aws.String(templateBody),
	}

	index := 0
	for key, value := range parameters {
		input.Parameters[index] = &cloudformation.Parameter{
			ParameterKey:   aws.String(key),
			ParameterValue: aws.String(value),
		}

		index += 1
	}

	out, err := client.CreateChangeSet(&input)

	if err != nil {
		return StackUpdate{}, err
	}

	return StackUpdate{
		Name:    *input.ChangeSetName,
		Id:      *out.Id,
		StackId: *out.StackId,
		client:  client,
	}, nil
}

func (update *StackUpdate) GetStatus() StackStatus {
	out, err := update.client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(update.StackId),
	})

	if err != nil {
		panic(err)
	}

	stack := out.Stacks[0]
	status := StackStatus(*stack.StackStatus)

	return status
}

func StackExists(sess *session.Session, stackName string) (bool, error) {
	client := cloudformation.New(sess)
	_, err := client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})

	// TODO: Check error type.
	if err != nil {
		return false, err
	}

	return true, nil
}
