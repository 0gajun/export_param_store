package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/urfave/cli"
)

var region string
var environment string
var identifier string

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"

	app.Name = "export_param_store"
	app.Usage = "Export AWS Parameter Store values as environment variables"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "env",
			Value:       "",
			Usage:       "Environment",
			Destination: &environment,
		},
		cli.StringFlag{
			Name:        "region",
			Value:       "",
			Usage:       "region",
			Destination: &region,
		},
		cli.StringFlag{
			Name:        "identifier",
			Value:       "",
			Usage:       "Identifier",
			Destination: &identifier,
		},
	}
	app.Action = run

	app.Run(os.Args)
}

func run(c *cli.Context) error {
	if err := validateArgs(); err != nil {
		return err
	}

	params, err := getParameters(c.Args(), region, environment, identifier)
	if err != nil {
		return err
	}

	printParametersAsExportForm(params)

	return nil
}

type Parameter struct {
	Name     string
	FullName string
	Value    string
}

type Parameters []Parameter

func parametersFrom(output *ssm.GetParametersOutput, prefix string) Parameters {
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

func getParameters(envs []string, region, environment, identifier string) (Parameters, error) {
	prefix := fmt.Sprintf("%s.%s.", environment, identifier)
	input := buildGetParameterQuery(prefix, envs)
	svc := newAwsService(region)
	output, err := svc.GetParameters(input)

	if err != nil {
		var nilSlice Parameters
		return nilSlice, cli.NewExitError(err.Error(), 1)
	}

	params := parametersFrom(output, prefix)

	return params, nil
}

func newAwsService(region string) *ssm.SSM {
	sess := session.Must(session.NewSession())
	svc := ssm.New(sess, &aws.Config{Region: aws.String(region)})
	return svc
}

func buildGetParameterQuery(prefix string, args []string) *ssm.GetParametersInput {
	input := ssm.GetParametersInput{}
	for _, envVarName := range args {
		loweredEnv := strings.ToLower(envVarName)
		paramName := fmt.Sprintf("%s%s", prefix, loweredEnv)
		input.Names = append(input.Names, &paramName)
	}
	input.SetWithDecryption(true)
	return &input
}

func printParametersAsExportForm(params Parameters) {
	for _, p := range params {
		name := strings.ToUpper(p.Name)
		fmt.Printf("export %s=%s\n", name, p.Value)
	}
}

func validateArgs() error {
	ok := true
	var errors cli.MultiError

	if len(region) == 0 {
		ok = false
		errors.Errors = append(errors.Errors, cli.NewExitError("region must be specified", 128))
	}
	if len(environment) == 0 {
		ok = false
		errors.Errors = append(errors.Errors, cli.NewExitError("environment must be specified", 128))
	}
	if len(identifier) == 0 {
		ok = false
		errors.Errors = append(errors.Errors, cli.NewExitError("identifier must be specified", 128))
	}

	if ok {
		return nil
	} else {
		return errors
	}
}
