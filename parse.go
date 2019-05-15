package main

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pkg/errors"
	"io/ioutil"
	"strings"
)

func parseParameterString(str string) (string, string) {
	split := strings.SplitN(str, "=", 2)
	key := split[0]
	value := ""

	if len(split) > 1 {
		value = split[1]
	}

	return key, value
}

func parseParameterFile(path string) (map[string]string, error) {
	result := make(map[string]string)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "read %s", path)
	}

	if strings.HasSuffix(path, ".json") {
		var params []cloudformation.Parameter
		err = json.Unmarshal(bytes, &params)
		if err != nil {
			return nil, errors.Wrapf(err, "malformed parameter file %s", path)
		}

		for _, param := range params {
			result[*param.ParameterKey] = *param.ParameterValue
		}
	} else {
		return nil, errors.Wrapf(err, "parameter file %s has an unknown extension", path)
	}

	return result, nil
}
