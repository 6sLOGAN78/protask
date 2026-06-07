package aws

import (
	"context"

	"github.com/6sLOGAN78/go-protask/internal/server"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWS struct {
	S3 *S3Client
}

func NewAWS(server *server.Server) (*AWS, error) {
	awsConfig := server.Config.AWS

	configOptions := []func(*config.LoadOptions) error{
		config.WithRegion(awsConfig.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				awsConfig.AccessKeyID,
				awsConfig.SecretAccessKey,
				"",
			),
		),
	}

	// Optional endpoint for S3-compatible providers
	if awsConfig.EndpointURL != "" {
		configOptions = append(
			configOptions,
			config.WithEndpointResolverWithOptions(
				aws.EndpointResolverWithOptionsFunc(
					func(service, region string,
						options ...interface{},
					) (aws.Endpoint, error) {

						if service == s3.ServiceID {
							return aws.Endpoint{
								URL:           awsConfig.EndpointURL,
								SigningRegion: awsConfig.Region,
							}, nil
						}

						return aws.Endpoint{}, &aws.EndpointNotFoundError{}
					},
				),
			),
		)
	}

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		configOptions...,
	)
	if err != nil {
		return nil, err
	}

	return &AWS{
		S3: NewS3Client(server, cfg),
	}, nil
}