package paramstore

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/urfave/cli"
)

const (
	requestChunkSize = 10
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
	svc := c.newAwsService()

	requestedEnvCount := len(envs)

	queries := c.buildGetParameterQueries(envs)
	queryCount := len(queries)

	ch := make(chan Parameters)
	defer close(ch)

	eg, _ := errgroup.WithContext(context.Background())
	for _, query := range queries {
		eg.Go(func() error {
			output, err := svc.GetParameters(query)
			if err != nil {
				return err
			}

			ch <- NewParameter(output, c.prefix)

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		var nilSlice Parameters
		return nilSlice, cli.NewExitError(err.Error(), 1)
	}

	result := Parameters{}
	for i := 0; i < queryCount; i++ {
		select {
		case params := <-ch:
			result = append(result, params...)
		}
	}

	if requestedEnvCount != len(result) {
		var nilSlice Parameters
		var fetchedEnvs = []string{}
		for _, param := range result {
			fetchedEnvs = append(fetchedEnvs, param.Name)
		}
		msg := fmt.Sprintf("Cannot fetch required params. Required params are %v. But got only %v", envs, fetchedEnvs)
		return nilSlice, cli.NewExitError(msg, 1)
	}

	return result, nil
}

func (c *Client) newAwsService() *ssm.SSM {
	sess := session.Must(session.NewSession())
	svc := ssm.New(sess, &aws.Config{Region: aws.String(c.region)})
	return svc
}

func (c *Client) buildGetParameterQueries(envs []string) []*ssm.GetParametersInput {
	paramNames := []string{}
	for _, envVarName := range envs {
		loweredEnv := strings.ToLower(envVarName)
		paramName := fmt.Sprintf("%s%s", c.prefix, loweredEnv)
		paramNames = append(paramNames, paramName)
	}

	inputs := []*ssm.GetParametersInput{}
	for _, chunk := range splitIntoChunks(paramNames, requestChunkSize) {
		input := &ssm.GetParametersInput{}
		input.Names = aws.StringSlice(chunk)
		input.SetWithDecryption(true)

		inputs = append(inputs, input)
	}

	return inputs
}

func splitIntoChunks(array []string, chunkSize int) [][]string {
	size := len(array)

	chunks := [][]string{}
	for fromIndex, toIndex := 0, 0; toIndex < size; fromIndex = toIndex {
		for ; toIndex-fromIndex < chunkSize && toIndex < size; toIndex++ {
		}
		chunks = append(chunks, array[fromIndex:toIndex])
	}

	return chunks
}

func buildPrefix(environment, identifier string) string {
	return fmt.Sprintf("%s.%s.", environment, identifier)
}
