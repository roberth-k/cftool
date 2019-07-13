package pprint

import (
	"github.com/aws/aws-sdk-go/service/sts"
	"io"
)

func Whoami(w io.Writer, region *string, id *sts.GetCallerIdentityOutput) {
	Field(w, "Account", *id.Account)
	Field(w, "Role", *id.Arn)

	if region != nil && *region != "" {
		Field(w, "Region", *region)
	}
}
