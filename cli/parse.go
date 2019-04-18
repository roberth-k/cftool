package main

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/tetratom/cfn-tool/cli/pprint"
	"io/ioutil"
	"strings"
)

func ParseParameterFromCommandLine(param string) cloudformation.Parameter {
	split := strings.SplitN(param, "=", 2)
	result := cloudformation.Parameter{}
	result.ParameterKey = aws.String(split[0])

	if len(split) > 1 {
		result.ParameterValue = aws.String(split[1])
	} else {
		result.ParameterValue = aws.String("")
	}

	return result
}

func ParseParameterFile(path string) (map[string]string, error) {
	result := make(map[string]string)
	bytes, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(path, ".json") {
		var params []cloudformation.Parameter
		err = json.Unmarshal(bytes, &params)

		if err != nil {
			pprint.UserErrorf("Malformed Parameter file %s.", path)
			return nil, err
		}

		for _, param := range params {
			result[*param.ParameterKey] = *param.ParameterValue
		}
	} else {
		return nil, errors.New("Parameter file " + path + " has an unknown extension")
	}

	return result, nil
}
