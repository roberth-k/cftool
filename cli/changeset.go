package main

import (
	"fmt"
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/tetratom/cfn-tool/cli/pprint"
)

func PPrintChangeSet(cs *cf.DescribeChangeSetOutput) {
	for _, change := range cs.Changes {
		// Spacing
		fmt.Println()

		if *change.Type != cf.ChangeTypeResource {
			pprint.Warningf("skipping unknown resource type: %s", *change.Type)
			continue
		}

		change := change.ResourceChange
		pprint.PrintChangeHeader(
			*change.Action,
			*change.ResourceType,
			*change.LogicalResourceId)

		if change.PhysicalResourceId != nil {
			fmt.Println(*change.PhysicalResourceId)
		}
	}
}
