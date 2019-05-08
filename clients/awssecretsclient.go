package clients

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// awsSecretsClient holds the AWS Client needed for integration
type awsSecretsClient struct {
	Client *secretsmanager.SecretsManager
}

// NewAWSSecretsClient provides a new SecretsManagerClient for using AWS secrets manager
func NewAWSSecretsClient(auth AWSSecretsAuth) (SecretsManagerClient, error) {
	cfg, err := auth.GetConfig()
	if err != nil {
		return nil, err
	}
	sess, err := session.NewSession(cfg)
	return &awsSecretsClient{
		Client: secretsmanager.New(sess),
	}, err
}

// GetSecrets will provide a slice of Secret type responses, for remote secrets located in AWS
func (client awsSecretsClient) GetSecrets(namespaces ...string) ([]*Secret, error) {
	awsSecrets := make([]*secretsmanager.SecretListEntry, 0)
	secretsList, err := client.Client.ListSecrets(&secretsmanager.ListSecretsInput{})
	if err != nil {
		return nil, err
	}
	awsSecrets = append(awsSecrets, secretsList.SecretList...)
	for {
		if secretsList.NextToken != nil {
			secretsList, err = client.Client.ListSecrets(&secretsmanager.ListSecretsInput{NextToken: secretsList.NextToken})
			if err != nil {
				return nil, err
			}
			awsSecrets = append(awsSecrets, secretsList.SecretList...)
		} else {
			break
		}
	}

	sc := make(chan *Secret)
	wg := &sync.WaitGroup{}
	for _, awsSecret := range awsSecrets {
		managed, paths := isManagedAWSSecret(awsSecret.Tags)
		if managed && paths != nil {
			wg.Add(1)
			go buildSecretFromAWSSecret(sc, wg, client.Client, *paths, *awsSecret.Name, namespaces...)
		}
	}

	go func() {
		wg.Wait()
		close(sc)
	}()

	secrets := make([]*Secret, 0)
	for secret := range sc {
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// GetSecret will retrieve a remote secret from AWS Secrets Manager
func (client awsSecretsClient) GetSecret(path string) (*Secret, error) {
	awsSecretValue, err := client.Client.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(path),
		VersionStage: aws.String("AWSCURRENT"),
	})
	if err != nil {
		return nil, err
	}
	secretData, err := buildAWSSecretData(awsSecretValue)
	if err != nil {
		return nil, err
	}
	return &Secret{Name: path, Data: secretData}, nil
}

// isManagedAWSSecret determines if an AWS secret is meant to be read by mimir
func isManagedAWSSecret(tags []*secretsmanager.Tag) (bool, *string) {
	managed := false
	paths := ""
	for _, tag := range tags {
		switch *tag.Key {
		case Managed:
			managed = *tag.Value == "true"
		case Paths:
			paths = *tag.Value
		}
	}
	if paths != "" {
		return managed, &paths
	}
	return managed, nil
}

// buildSecretFromAWSSecret calls AWS to get the value of a secret found to be managed by mimir
func buildSecretFromAWSSecret(sc chan<- *Secret, wg *sync.WaitGroup, client *secretsmanager.SecretsManager, paths, secretName string, namespaces ...string) {
	defer wg.Done()
	awsSecretValue, err := client.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	})
	if err != nil {
		log.Println(err.Error())
	}
	buildSecretFromAWSSecretValue(sc, awsSecretValue, paths, namespaces...)
}

// buildSecretFromAWSSecretValue constructs a Secret type of response from an AWS secret value retrieved
func buildSecretFromAWSSecretValue(sc chan<- *Secret, awsSecretValue *secretsmanager.GetSecretValueOutput, paths string, namespaces ...string) {
	secretData, err := buildAWSSecretData(awsSecretValue)
	if err != nil {
		log.Println(err.Error())
	}
	splitPaths := strings.Split(paths, "+")
	for _, splitPath := range splitPaths {
		splitK8SPath := strings.Split(splitPath, "/")
		nsExists := func() bool {
			for _, namespace := range namespaces {
				if namespace == splitK8SPath[0] {
					return true
				}
			}
			return false
		}()
		if nsExists {
			sc <- &Secret{
				Name:      splitK8SPath[1],
				Namespace: splitK8SPath[0],
				Data:      secretData,
			}
		}
	}
}

// buildAWSSecretData converts the AWS secret data into a k8s friendly type for later use
func buildAWSSecretData(awsSecret *secretsmanager.GetSecretValueOutput) (map[string]string, error) {
	data := make(map[string]string)
	if awsSecret.SecretString != nil {
		secretBytes := []byte(*awsSecret.SecretString)
		err := json.Unmarshal(secretBytes, &data)
		return data, err
	}
	if awsSecret.SecretBinary != nil {
		secretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(awsSecret.SecretBinary)))
		l, err := base64.StdEncoding.Decode(secretBytes, awsSecret.SecretBinary)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(secretBytes[:l], &data)
		return data, err
	}
	return nil, fmt.Errorf("Could not find any secret data for %s", *awsSecret.Name)
}
