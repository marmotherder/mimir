package clients

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	mgmtkv "github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2018-02-14/keyvault"
)

func strPointer(s string) *string {
	return &s
}

type mockBackendMgmtClient struct{}
type mockBackendClient struct{}

func (client mockBackendMgmtClient) ListBySubscription(ctx context.Context, top *int32) (result mgmtkv.VaultListResultPage, err error) {
	run := false
	page := mgmtkv.NewVaultListResultPage(func(ctx context.Context, results mgmtkv.VaultListResult) (mgmtkv.VaultListResult, error) {
		if !run {
			run = true
			return mgmtkv.VaultListResult{
				Value: &[]mgmtkv.Vault{
					mgmtkv.Vault{
						Name: strPointer("mock"),
						ID:   strPointer("mock"),
						Properties: &mgmtkv.VaultProperties{
							VaultURI: strPointer("https://mock.mock"),
						},
						Tags: map[string]*string{
							"mimir-managed": strPointer("true"),
							"mimir-paths":   strPointer("mock/mock+mock1/mock1+mock2/mock2"),
						},
					},
				},
			}, nil
		}
		return mgmtkv.VaultListResult{}, nil
	})
	page.Next()
	return page, nil
}

func (client mockBackendClient) GetSecrets(ctx context.Context, vaultBaseURL string, maxresults *int32) (result keyvault.SecretListResultPage, err error) {
	run := false
	page := keyvault.NewSecretListResultPage(func(context.Context, keyvault.SecretListResult) (keyvault.SecretListResult, error) {
		if !run {
			run = true
			return keyvault.SecretListResult{
				Value: &[]keyvault.SecretItem{
					keyvault.SecretItem{
						ID: strPointer("mock/mock"),
					},
				},
			}, nil
		}
		return keyvault.SecretListResult{}, nil
	})
	page.Next()
	return page, nil
}

func (client mockBackendClient) GetSecret(ctx context.Context, vaultBaseURL string, secretName string, secretVersion string) (result keyvault.SecretBundle, err error) {
	return keyvault.SecretBundle{
		ID:    strPointer("mock/mock"),
		Value: strPointer("mock"),
	}, nil
}

func TestAzureGetSecret(t *testing.T) {
	client := azureSecretsClient{
		mockBackendMgmtClient{},
		mockBackendClient{},
	}

	secret, err := client.GetSecret("mock")
	if err != nil {
		t.Error(err.Error())
	}

	if secret.Name != "mock" {
		t.Error("Did not load with expected secret name")
	}

	if value, ok := secret.Data["mock"]; ok {
		if value != "mock" {
			t.Error("Data value not as expected")
		}
	} else {
		t.Error("Did not locate the expected data key")
	}
}

func TestAzureGetSecrets(t *testing.T) {
	client := azureSecretsClient{
		mockBackendMgmtClient{},
		mockBackendClient{},
	}

	secrets, err := client.GetSecrets("mock", "mock1")
	if err != nil {
		t.Error(err.Error())
	}

	if len(secrets) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(secrets))
	}

	if secrets[0].Name != "mock" ||
		secrets[0].Namespace != "mock" {
		t.Error("Name or namespace of the first secret isn't right")
	}

	if secrets[1].Name != "mock1" ||
		secrets[1].Namespace != "mock1" {
		t.Error("Name or namespace of the second secret isn't right")
	}

	if value, ok := secrets[0].Data["mock"]; ok {
		if value != "mock" {
			t.Error("Data value not as expected")
		}
	} else {
		t.Error("Did not locate the expected data key")
	}
}
