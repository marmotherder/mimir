package clients

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	mgmtkv "github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2018-02-14/keyvault"
)

type azureSecretsClient struct {
	MgmtClient backendMgmtClient
	Client     backendClient
}

type backendMgmtClient interface {
	ListBySubscription(ctx context.Context, top *int32) (result mgmtkv.VaultListResultPage, err error)
}

type backendClient interface {
	GetSecrets(ctx context.Context, vaultBaseURL string, maxresults *int32) (result keyvault.SecretListResultPage, err error)
	GetSecret(ctx context.Context, vaultBaseURL string, secretName string, secretVersion string) (result keyvault.SecretBundle, err error)
}

// NewAzureKeyVaultClient load a new instance of a secrets client for Azure Key Vault
func NewAzureKeyVaultClient(auth AzureKeyVaultAuth, subscriptionID ...string) (SecretsManagerClient, error) {
	subID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) > 0 && subscriptionID[0] != "" {
		subID = subscriptionID[0]
	}

	if subID == "" {
		return nil, errors.New("Failed to find a valid subscription ID for azure")
	}

	mgmtauthorizer, err := auth.GetMgmtAuth()
	if err != nil {
		return nil, err
	}
	if mgmtauthorizer == nil {
		return nil, errors.New("Failed to correctly load authorizer for azure")
	}

	mgmtc := mgmtkv.NewVaultsClient(subID)
	mgmtc.Authorizer = *mgmtauthorizer

	authorizer, err := auth.GetAuth()
	if err != nil {
		return nil, err
	}
	if authorizer == nil {
		return nil, errors.New("Failed to correctly load authorizer for azure")
	}

	c := keyvault.New()
	c.Authorizer = *authorizer

	return &azureSecretsClient{
		&mgmtc,
		&c,
	}, err
}

// GetSecrets returns a list of k8s compatible secrets as loaded from secrets within an Azure Key Vault
func (client azureSecretsClient) GetSecrets(namespaces ...string) ([]*Secret, error) {
	vaults, err := client.MgmtClient.ListBySubscription(context.Background(), nil)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	var secrets []*Secret
	for {
		for _, vault := range vaults.Values() {
			if vault.Name != nil {
				if managed, paths := isManagedVault(vault.Tags); managed && paths != nil {
					vaultSecrets, err := client.Client.GetSecrets(context.Background(), *vault.Properties.VaultURI, nil)
					if err != nil {
						return nil, err
					}
					secret, err := client.loadSecret(*vault.Properties.VaultURI, vaultSecrets)
					for _, path := range strings.Split(*paths, "+") {
						splitPath := strings.Split(path, "/")
						ns := func() bool {
							for _, namespace := range namespaces {
								if namespace == splitPath[0] {
									return true
								}
							}
							return false
						}
						if ns() {
							secrets = append(secrets, &Secret{
								Name:      splitPath[1],
								Namespace: splitPath[0],
								Data:      secret.Data,
							})
						}
					}
				}
			}
		}
		if vaults.NotDone() {
			if err := vaults.Next(); err != nil {
				return nil, err
			}
		} else {
			break
		}
	}
	return secrets, nil
}

// GetSecret loads a single k8s compatible secret from an Azure Key Vault secret data
func (client azureSecretsClient) GetSecret(path string) (*Secret, error) {
	vaults, err := client.MgmtClient.ListBySubscription(context.Background(), nil)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	var r mgmtkv.Vault
	for {
		for _, vault := range vaults.Values() {
			if vault.Name != nil && *vault.Name == path {
				managed, _ := isManagedVault(vault.Tags)
				if managed {
					r = vault
				}
			}
		}
		if vaults.NotDone() {
			if err := vaults.Next(); err != nil {
				return nil, err
			}
		} else {
			break
		}
	}
	if r.ID == nil {
		return nil, errors.New("Failed to find a key vault at the selected path")
	}
	if r.Properties.VaultURI == nil {
		return nil, fmt.Errorf("Vault with id: %s has no associated URI", *r.ID)
	}
	secretItems, err := client.Client.GetSecrets(context.Background(), *r.Properties.VaultURI, nil)
	if err != nil {
		return nil, err
	}
	secret, err := client.loadSecret(*r.Properties.VaultURI, secretItems)
	if err != nil {
		return nil, err
	}
	secret.Name = path

	return secret, nil
}

func (client azureSecretsClient) loadSecret(vaultURI string, secrets keyvault.SecretListResultPage) (*Secret, error) {
	result := Secret{
		Data: make(map[string]string),
	}
	for {
		for _, secret := range secrets.Values() {
			if secret.ID == nil {
				continue
			}
			splitPath := strings.Split(*secret.ID, "/")
			sr, err := client.Client.GetSecret(context.Background(), vaultURI, splitPath[len(splitPath)-1], "")
			if err != nil {
				return nil, err
			}
			if sr.Value != nil {
				result.Data[splitPath[len(splitPath)-1]] = *sr.Value
			}
		}
		if secrets.NotDone() {
			if err := secrets.Next(); err != nil {
				return nil, err
			}
		} else {
			break
		}
	}
	return &result, nil
}

func isManagedVault(tags map[string]*string) (bool, *string) {
	managed := false
	var paths *string
	for key, value := range tags {
		switch key {
		case Managed:
			if value != nil {
				managed = *value == "true"
			}
		case Paths:
			paths = value
		}
	}
	if paths != nil {
		return managed, paths
	}
	return managed, nil
}
