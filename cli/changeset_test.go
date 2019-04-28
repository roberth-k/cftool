package main

import (
	"github.com/aws/aws-sdk-go/aws"
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stretchr/testify/require"
	"github.com/tetratom/cfn-tool/cli/pprint"
	"strings"
	"testing"
)

func TestPPrintChangeSet(t *testing.T) {
	pprint.DisableColor()
	defer pprint.EnableColor()
	b := strings.Builder{}
	pprint.SetWriter(&b)

	tests := []struct {
		ChangeSet cf.DescribeChangeSetOutput
		Expect    string
	}{
		{
			cf.DescribeChangeSetOutput{
				Changes: []*cf.Change{
					{
						Type: aws.String("Resource"),
						ResourceChange: &cf.ResourceChange{
							Replacement:       aws.String("False"),
							ResourceType:      aws.String("AWS::Resource"),
							Action:            aws.String(cf.ChangeActionAdd),
							LogicalResourceId: aws.String("MyResource"),
						},
					},
					{
						Type: aws.String("Resource"),
						ResourceChange: &cf.ResourceChange{
							Replacement:       aws.String("False"),
							ResourceType:      aws.String("AWS::ModifiedResource"),
							Action:            aws.String(cf.ChangeActionModify),
							LogicalResourceId: aws.String("MyResource"),
							Details: []*cf.ResourceChangeDetail{
								{
									CausingEntity: aws.String("MyProp"),
									Evaluation:    aws.String(cf.EvaluationTypeStatic),
									ChangeSource:  aws.String(cf.ChangeSourceResourceAttribute),
									Target: &cf.ResourceTargetDefinition{
										RequiresRecreation: aws.String(cf.RequiresRecreationConditionally),
										Attribute:          aws.String("MyAtt"),
										Name:               aws.String("MyProperty"),
									},
								},
							},
						},
					},
					{
						Type: aws.String("Resource"),
						ResourceChange: &cf.ResourceChange{
							Replacement:        aws.String("True"),
							ResourceType:       aws.String("AWS::ReplacedResource"),
							Action:             aws.String(cf.ChangeActionModify),
							LogicalResourceId:  aws.String("MyResource"),
							PhysicalResourceId: aws.String("PhysicalId"),
						},
					},
				},
			},
			`
+ AWS::Resource MyResource

~ AWS::ModifiedResource MyResource
    Change: MyAtt.MyProperty <- !GetAtt MyProp (conditional replacement)

- AWS::ReplacedResource MyResource
+ AWS::ReplacedResource MyResource
PhysicalId
`,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			b.Reset()
			PPrintChangeSet(&test.ChangeSet)
			require.Equal(t, test.Expect, b.String())
		})
	}
}
