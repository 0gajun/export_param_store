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
var envVarNames cli.StringSlice

func main() {
	app := cli.NewApp()

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
		cli.StringSliceFlag{
			Name:  "env_var",
			Usage: "Environment Variable Name",
			Value: &envVarNames,
		},
	}
	app.Action = exportParamStoreAsEnvVar

	app.Run(os.Args)
}

func exportParamStoreAsEnvVar(c *cli.Context) error {
	if err := validateArgs(); err != nil {
		return err
	}

	svc := newAwsService(region)

	prefix := fmt.Sprintf("%s.%s.", environment, identifier)

	input := buildGetParameterQuery(prefix, c.Args())

	output, err := svc.GetParameters(input)

	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	for _, p := range output.Parameters {
		name := strings.ToUpper(strings.Replace(*p.Name, prefix, "", 1))
		fmt.Printf("export %s=%s\n", name, *p.Value)
	}

	return nil
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

func newAwsService(region string) *ssm.SSM {
	sess := session.Must(session.NewSession())
	svc := ssm.New(sess, &aws.Config{Region: aws.String(region)})
	return svc
}
