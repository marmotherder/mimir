package clients

import (
	"testing"

	core_v1 "k8s.io/api/core/v1"
)

func TestGetManagedSecrets(t *testing.T) {
	secret := core_v1.Secret{}
	secret.Annotations = map[string]string{
		Managed: "true",
		Source:  string(AWS),
	}
	if len(getManagedSecrets([]core_v1.Secret{secret}, AWS)) < 1 {
		t.Error("Secret did not show as managed")
	}
}
