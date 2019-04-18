package cfn

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/google/uuid"
)

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
