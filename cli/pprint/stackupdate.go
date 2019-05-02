package pprint

import (
	"fmt"
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	"io"
)

func str(s *string, def string) string {
	if s == nil {
		return def
	}

	return *s
}

func ChangeHeader(w io.Writer, action string, resourceType string, logicalResourceId string) {
	symbol := "???"
	col := ColModify

	switch action {
	case cf.ChangeActionRemove:
		symbol = "-"
		col = ColRemove

	case cf.ChangeActionModify:
		symbol = "~"
		col = ColModify

	case cf.ChangeActionAdd:
		symbol = "+"
		col = ColAdd
	}

	col.Fprintf(w, "%s %s", symbol, resourceType)
	ColLogicalId.Fprintf(w, " %s", logicalResourceId)
	fmt.Fprintf(w, "\n")
}

func ChangeSet(w io.Writer, cs *cf.DescribeChangeSetOutput) {
	for _, change := range cs.Changes {
		fmt.Fprintf(w, "\n") // Spacing.

		if *change.Type != cf.ChangeTypeResource {
			ColWarning.Fprintf(w, "skipping unknown resource type: %s", *change.Type)
			continue
		}

		change := change.ResourceChange
		replacement := str(change.Replacement, "")

		// Display change type. Replacements show up as a remove-and-add.
		if replacement == cf.ReplacementTrue {
			ChangeHeader(
				w,
				cf.ChangeActionRemove,
				*change.ResourceType,
				*change.LogicalResourceId)
			ChangeHeader(
				w,
				cf.ChangeActionAdd,
				*change.ResourceType,
				*change.LogicalResourceId)
		} else {
			ChangeHeader(
				w,
				*change.Action,
				*change.ResourceType,
				*change.LogicalResourceId)
		}

		// The physical ID is just a line.
		if change.PhysicalResourceId != nil {
			Field(w, " Resource", *change.PhysicalResourceId)
		}

		for _, detail := range change.Details {
			ChangeSetDetail(w, detail)
		}
	}
}

func ChangeSetDetail(w io.Writer, detail *cf.ResourceChangeDetail) {
	changeSource := str(detail.ChangeSource, "")
	targetAttribute := str(detail.Target.Attribute, "")
	targetPropertyName := str(detail.Target.Name, "")
	targetRequiresRecreation := str(detail.Target.RequiresRecreation, "")
	evaluation := str(detail.Evaluation, "")
	causingEntity := str(detail.CausingEntity, "")

	BeginField(w, "   Change")

	fmt.Fprintf(w, targetAttribute)
	if targetPropertyName != "" {
		fmt.Fprintf(w, ".%s", targetPropertyName)
	}

	if evaluation == cf.EvaluationTypeDynamic {
		fmt.Fprintf(w, " <~")
	} else {
		fmt.Fprintf(w, " <-")
	}

	switch changeSource {
	case cf.ChangeSourceResourceReference, cf.ChangeSourceParameterReference:
		fmt.Fprintf(w, " !Ref")
		if causingEntity != "" {
			fmt.Fprintf(w, " %s", causingEntity)
		}
	case cf.ChangeSourceResourceAttribute:
		fmt.Fprintf(w, " !GetAtt")
		if causingEntity != "" {
			fmt.Fprintf(w, " %s", causingEntity)
		}
	case cf.ChangeSourceDirectModification, cf.ChangeSourceAutomatic:
		fmt.Fprintf(w, " ...")
	default:
		fmt.Fprintf(w, " ??? unknown change source \"%s\"", changeSource)

	}

	commentCount := 0
	switch changeSource {
	case cf.ChangeSourceDirectModification:
		fmt.Fprintf(w, " (direct modification")
		commentCount += 1

	case cf.ChangeSourceAutomatic:
		fmt.Fprintf(w, " (automatic")
		commentCount += 1
	}

	switch targetRequiresRecreation {
	case cf.RequiresRecreationConditionally:
		if commentCount == 0 {
			fmt.Fprintf(w, " (")
		} else {
			fmt.Fprintf(w, ", ")
		}

		ColWarning.Fprintf(w, "conditional replacement")
		commentCount += 1

	case cf.RequiresRecreationAlways:
		if commentCount == 0 {
			fmt.Fprintf(w, " (")
		} else {
			fmt.Fprintf(w, ", ")
		}

		ColError.Fprintf(w, "always replace")
		commentCount += 1
	}

	if commentCount > 0 {
		fmt.Fprintf(w, ")")
	}

	fmt.Fprintf(w, "\n")
}

func StackEvent(w io.Writer, event *cf.StackEvent) {
	ColError.Fprintf(w, "Error! %s", *event.ResourceType)
	ColLogicalId.Fprintf(w, " %s", *event.LogicalResourceId)
	fmt.Fprintf(w, ": %s\n", str(event.ResourceStatusReason, "???"))
}

func StackOutput(w io.Writer, output *cf.Output) {
	Field(w, *output.OutputKey, *output.OutputValue)
}
