package clients

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

type AWSSecretsAuth interface {
	GetConfig(region string) (*aws.Config, error)
}

type AWSIAMAuth struct {
	Region string
}

func (auth AWSIAMAuth) GetConfig() (*aws.Config, error) {
	return aws.NewConfig().WithRegion(auth.Region), nil
}

type AWSStaticCredentialsAuth struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

func (auth AWSStaticCredentialsAuth) GetConfig() (*aws.Config, error) {
	if auth.AccessKeyID == "" {
		return nil, errors.New("Requires an access key id")
	}
	if auth.SecretAccessKey == "" {
		return nil, errors.New("Requires a secret access key")
	}

	return aws.NewConfig().WithRegion(auth.Region).WithCredentials(credentials.NewStaticCredentials(auth.AccessKeyID, auth.SecretAccessKey, "")), nil
}
