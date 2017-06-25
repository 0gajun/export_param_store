package paramstore

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/urfave/cli"
)

type Client struct {
	region      string
	environment string
	identifier  string
	prefix      string
}

func NewClient(region, environment, identifier string) Client {
	return Client{
		region:      region,
		environment: environment,
		identifier:  identifier,
		prefix:      buildPrefix(environment, identifier),
	}
}

func (c *Client) GetParameters(envs []string) (Parameters, error) {
	query := c.buildGetParameterQuery(envs)
	svc := c.newAwsService()
	output, err := svc.GetParameters(query)

	if err != nil {
		var nilSlice Parameters
		return nilSlice, cli.NewExitError(err.Error(), 1)
	}

	params := NewParameter(output, c.prefix)

	return params, nil
}

func (c *Client) newAwsService() *ssm.SSM {
	sess := session.Must(session.NewSession())
	svc := ssm.New(sess, &aws.Config{Region: aws.String(c.region)})
	return svc
}

func (c *Client) buildGetParameterQuery(envs []string) *ssm.GetParametersInput {
	input := ssm.GetParametersInput{}
	for _, envVarName := range envs {
		loweredEnv := strings.ToLower(envVarName)
		paramName := fmt.Sprintf("%s%s", c.prefix, loweredEnv)
		input.Names = append(input.Names, &paramName)
	}
	input.SetWithDecryption(true)
	return &input
}

func buildPrefix(environment, identifier string) string {
	return fmt.Sprintf("%s.%s.", environment, identifier)
}
