package clients

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func TestIsManagedAWSSecret(t *testing.T) {
	validTags := []*secretsmanager.Tag{
		&secretsmanager.Tag{Key: aws.String(Managed), Value: aws.String("true")},
		&secretsmanager.Tag{Key: aws.String(Paths), Value: aws.String("mockns1/mocksec1+mockns2/mocksec1")},
	}

	managed, paths := isManagedAWSSecret(validTags)
	if !managed || paths == nil {
		t.Error("Did not come back as fully managed")
	}

	noPathTags := []*secretsmanager.Tag{
		&secretsmanager.Tag{Key: aws.String(Managed), Value: aws.String("true")},
	}

	managed, paths = isManagedAWSSecret(noPathTags)
	if managed && paths != nil {
		t.Error("Got paths unexpectedly")
	}
}

func TestBuildSecretFromAWSSecretValue(t *testing.T) {
	sc := make(chan *Secret)
	awsSecretValue := &secretsmanager.GetSecretValueOutput{
		SecretString: aws.String("{\"mock\": \"mock\"}"),
	}

	results := make([]*Secret, 0)
	go func() {
		for secret := range sc {
			results = append(results, secret)
		}
	}()

	buildSecretFromAWSSecretValue(sc, awsSecretValue, "mockns1/mock1+mockns2/mock1", "mockns1", "mockns2")
	close(sc)

	if len(results) < 2 {
		t.Error("Did not get all expected secrets back")
	}
}
