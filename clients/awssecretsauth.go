package clients

import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// AWSSecretsAuth interface provides a common function set to authenticate with AWS from mimir
type AWSSecretsAuth interface {
	SetRegion(region string) error
	GetConfig() (*aws.Config, error)
}

// AWSRegion is a common struct for setting the AWS region to use
type AWSRegion struct {
	Region string
}

// SetRegion is a common function for all authentication structs to set the region to use.
func (auth *AWSRegion) SetRegion(region string) error {
	auth.Region = region
	if auth.Region == "" {
		if envRegion, found := os.LookupEnv("AWS_REGION"); found {
			auth.Region = envRegion
		}
		return errors.New("AWS_REGION environment variable not set")
	}
	return nil
}

// AWSIAMAuth contains auth information for using IAM to authenticate
type AWSIAMAuth struct {
	AWSRegion
}

// GetConfig will load the AWS config for IAM Auth
func (auth AWSIAMAuth) GetConfig() (*aws.Config, error) {
	return aws.NewConfig().WithRegion(auth.Region), nil
}

// AWSStaticCredentialsAuth contains auth information for using static credentials to authenticate
type AWSStaticCredentialsAuth struct {
	AccessKeyID     string
	SecretAccessKey string
	AWSRegion
}

// GetConfig will load the AWS config for static credentials
func (auth AWSStaticCredentialsAuth) GetConfig() (*aws.Config, error) {
	if auth.AccessKeyID == "" {
		return nil, errors.New("Requires an access key id")
	}
	if auth.SecretAccessKey == "" {
		return nil, errors.New("Requires a secret access key")
	}

	return aws.NewConfig().WithRegion(auth.Region).WithCredentials(credentials.NewStaticCredentials(auth.AccessKeyID, auth.SecretAccessKey, "")), nil
}

// AWSEnvironmentAuth contains auth information for using environment variables to authenticate
type AWSEnvironmentAuth struct {
	AWSRegion
}

// GetConfig will load the AWS config for environment variables
func (auth AWSEnvironmentAuth) GetConfig() (*aws.Config, error) {
	_, foundID := os.LookupEnv("AWS_ACCESS_KEY_ID")
	if !foundID {
		return nil, errors.New("AWS_ACCESS_KEY_ID environment variable not set")
	}
	_, foundSecret := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !foundSecret {
		return nil, errors.New("AWS_SECRET_ACCESS_KEY environment variable not set")
	}

	return aws.NewConfig().WithRegion(auth.Region).WithCredentials(credentials.NewEnvCredentials()), nil
}

// AWSSharedCredentialsAuth contains auth information for using shared credentials file to authenticate
type AWSSharedCredentialsAuth struct {
	Path    string
	Profile string
	AWSRegion
}

// GetConfig will load the AWS config for shared credentials file
func (auth AWSSharedCredentialsAuth) GetConfig() (*aws.Config, error) {
	path := auth.Path
	if path == "" {
		path = fmt.Sprintf("%s/.aws/credentials", getHomeDir())
	}
	profile := auth.Profile
	if profile == "" {
		profile = "default"
	}

	return aws.NewConfig().WithRegion(auth.Region).WithCredentials(credentials.NewSharedCredentials(path, profile)), nil
}
