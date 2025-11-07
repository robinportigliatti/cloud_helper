package rds

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type DescribeDBParametersResult struct {
	Parameters []Parameter `json:"Parameters"`
}

type Parameter struct {
	ParameterName        string `json:"ParameterName"`
	Description          string `json:"Description"`
	Source               string `json:"Source"`
	ApplyType            string `json:"ApplyType"`
	DataType             string `json:"DataType"`
	IsModifiable         bool   `json:"IsModifiable"`
	ApplyMethod          string `json:"ApplyMethod"`
	ParameterValue       string `json:"ParameterValue,omitempty"`
	AllowedValues        string `json:"AllowedValues,omitempty"`
	MinimumEngineVersion string `json:"MinimumEngineVersion,omitempty"`
}

func (d *DescribeDBParametersResult) GetParameterValueByParameterName(parameterName string) (string, error) {
	idx := slices.IndexFunc(d.Parameters, func(c Parameter) bool { return c.ParameterName == parameterName })

	if idx == -1 {
		return "", fmt.Errorf("IndexFunc: parameter %s not found", parameterName)
	}

	return d.Parameters[idx].ParameterValue, nil
}
