package main

import (
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/tetratom/cfn-tool/cli/pprint"
)

func PPrintChangeSet(cs *cf.DescribeChangeSetOutput) {
	for i, change := range cs.Changes {
		if i > 0 {
			// Spacing
			pprint.Write("")
		}

		if *change.Type != cf.ChangeTypeResource {
			pprint.Warningf("skipping unknown resource type: %s", *change.Type)
			continue
		}

		change := change.ResourceChange
		replacement := str(change.Replacement)

		// Display change type. Replacements show up as a remove-and-add.
		if replacement == cf.ReplacementTrue {
			pprint.PrintChangeHeader(
				cf.ChangeActionRemove,
				*change.ResourceType,
				*change.LogicalResourceId)
			pprint.PrintChangeHeader(
				cf.ChangeActionAdd,
				*change.ResourceType,
				*change.LogicalResourceId)
		} else {
			pprint.PrintChangeHeader(
				*change.Action,
				*change.ResourceType,
				*change.LogicalResourceId)
		}

		// The physical ID is just a line.
		if change.PhysicalResourceId != nil {
			pprint.Printf("%s\n", *change.PhysicalResourceId)
		}

		for _, detail := range change.Details {
			PPrintChangeSetDetail(detail)
		}
	}
}

func str(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func PPrintChangeSetDetail(detail *cf.ResourceChangeDetail) {
	changeSource := str(detail.ChangeSource)
	targetAttribute := str(detail.Target.Attribute)
	targetPropertyName := str(detail.Target.Name)
	targetRequiresRecreation := str(detail.Target.RequiresRecreation)
	evaluation := str(detail.Evaluation)
	causingEntity := str(detail.CausingEntity)

	pprint.StartField("    Change")

	pprint.Printf(targetAttribute)
	if targetPropertyName != "" {
		pprint.Printf(".%s", targetPropertyName)
	}

	if evaluation == cf.EvaluationTypeDynamic {
		pprint.Printf(" <~")
	} else {
		pprint.Printf(" <-")
	}

	switch changeSource {
	case cf.ChangeSourceResourceReference, cf.ChangeSourceParameterReference:
		pprint.Printf(" !Ref")
		if causingEntity != "" {
			pprint.Printf(" %s", causingEntity)
		}
	case cf.ChangeSourceResourceAttribute:
		pprint.Printf(" !GetAtt")
		if causingEntity != "" {
			pprint.Printf(" %s", causingEntity)
		}
	case cf.ChangeSourceDirectModification, cf.ChangeSourceAutomatic:
		pprint.Printf(" ...")
	default:
		pprint.Printf(" ??? unknown change source \"%s\"", changeSource)

	}

	commentCount := 0
	switch changeSource {
	case cf.ChangeSourceDirectModification:
		pprint.Printf(" (direct modification")
		commentCount += 1

	case cf.ChangeSourceAutomatic:
		pprint.Printf(" (automatic")
		commentCount += 1
	}

	switch targetRequiresRecreation {
	case cf.RequiresRecreationConditionally:
		if commentCount == 0 {
			pprint.Printf(" (")
		} else {
			pprint.Printf(", ")
		}

		pprint.Yellowf("conditional replacement")
		commentCount += 1

	case cf.RequiresRecreationAlways:
		if commentCount == 0 {
			pprint.Printf(" (")
		} else {
			pprint.Printf(", ")
		}

		pprint.Redf("always replace")
		commentCount += 1
	}

	if commentCount > 0 {
		pprint.Printf(")")
	}

	pprint.Printf("\n")
}
