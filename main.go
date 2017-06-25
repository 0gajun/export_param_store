package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/0gajun/export_param_store/paramstore"
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

	client := paramstore.NewClient(region, environment, identifier)

	params, err := client.GetParameters(c.Args())
	if err != nil {
		return err
	}

	for _, p := range params {
		fmt.Println(p.GetAsExportForm())
	}

	return nil
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
