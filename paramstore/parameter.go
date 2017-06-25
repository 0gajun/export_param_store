package paramstore

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ssm"
)

type Parameters []Parameter

type Parameter struct {
	Name     string
	FullName string
	Value    string
}

func NewParameter(output *ssm.GetParametersOutput, prefix string) Parameters {
	var params Parameters
	for _, p := range output.Parameters {
		name := removePrefix(*p.Name, prefix)
		params = append(params, Parameter{Name: name, FullName: *p.Name, Value: *p.Value})
	}
	return params
}

func removePrefix(input string, prefix string) string {
	return strings.Replace(input, prefix, "", 1)
}

func (p *Parameter) GetAsExportForm() string {
	name := strings.ToUpper(p.Name)
	return fmt.Sprintf("export %s=%s\n", name, p.Value)
}
