package secrets

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

const (
	awsOption = "aws"
)

// AWSFlags defines global (but hidden) CLI flags. The purpose of these
// CLI flags is to initialize the AWS SSM Parameter Store provider via
// environment variables and/or the application's configuration file.
func AWSFlags(configFilePath altsrc.StringSourcer) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name: "secrets-aws-region",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("AWS_REGION"),
				toml.TOML("secrets.aws.region", configFilePath),
			),
			Hidden: true,
		},
		&cli.StringFlag{
			Name: "secrets-aws-kms-key-id",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("AWS_KMS_KEY_ID"),
				toml.TOML("secrets.aws.kms_key_id", configFilePath),
			),
			Hidden: true,
		},
	}
}

type awsProvider struct {
	keyID  *string
	client *ssm.Client
}

func newAWSProvider(ctx context.Context, cmd *cli.Command) (Manager, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cmd.String("secrets-aws-region")))
	if err != nil {
		return nil, err
	}

	return &awsProvider{
		keyID:  aws.String(cmd.String("secrets-aws-kms-key-id")),
		client: ssm.NewFromConfig(cfg),
	}, nil
}

// Set value size limit is 4 KiB, according to this link:
// https://docs.aws.amazon.com/systems-manager/latest/userguide/parameter-store-advanced-parameters.html.
func (p *awsProvider) Set(ctx context.Context, key, value string) error {
	_, err := p.client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      aws.String("/" + key),
		Value:     aws.String(value),
		Type:      types.ParameterTypeSecureString,
		KeyId:     p.keyID,
		Overwrite: aws.Bool(true),
	})
	return err
}

func (p *awsProvider) Get(ctx context.Context, key string) (string, error) {
	out, err := p.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String("/" + key),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		var pnf *types.ParameterNotFound
		if errors.As(err, &pnf) {
			return "", nil
		}
		return "", err
	}

	return aws.ToString(out.Parameter.Value), nil
}

func (p *awsProvider) Delete(ctx context.Context, key string) error {
	_, err := p.client.DeleteParameter(ctx, &ssm.DeleteParameterInput{Name: aws.String("/" + key)})
	return err
}
