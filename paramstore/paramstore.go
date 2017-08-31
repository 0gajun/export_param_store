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

	queries := c.buildGetParameterQueries(envs)
	queryCount := len(queries)

	ch := make(chan Parameters)
	eg, _ := errgroup.WithContext(context.Background())

	for i := 0; i < queryCount; i++ {
		query := queries[i]

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
	var fromIndex = 0
	var toIndex = 0
	size := len(array)

	chunks := [][]string{}
	for {
		for ; toIndex-fromIndex <= chunkSize && toIndex < size; toIndex++ {
		}

		chunks = append(chunks, array[fromIndex:toIndex])

		if toIndex > size-1 {
			break
		} else {
			toIndex = fromIndex
		}
	}

	return chunks
}

func buildPrefix(environment, identifier string) string {
	return fmt.Sprintf("%s.%s.", environment, identifier)
}
