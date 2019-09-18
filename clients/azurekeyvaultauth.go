package clients

import (
	kvauth "github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

// AzureKeyVaultAuth is a generic interface for
// authenticating against Azure
type AzureKeyVaultAuth interface {
	GetMgmtAuth() (*autorest.Authorizer, error)
	GetAuth() (*autorest.Authorizer, error)
}

// AzureKeyVaultEnvironmentAuth is for authentication
// using environment credentials
type AzureKeyVaultEnvironmentAuth struct{}

// GetMgmtAuth loads an Azure authorizer using environment variables for the management
// layer
func (va AzureKeyVaultEnvironmentAuth) GetMgmtAuth() (*autorest.Authorizer, error) {
	at, err := auth.NewAuthorizerFromEnvironment()
	return &at, err
}

// GetAuth loads an Azure authorizer using environment variables for the Key Vault
// component specifically
func (va AzureKeyVaultEnvironmentAuth) GetAuth() (*autorest.Authorizer, error) {
	at, err := kvauth.NewAuthorizerFromEnvironment()
	return &at, err
}

// AzureKeyVaultFileAuth is for authentication
// using a credentials file
type AzureKeyVaultFileAuth struct {
	BaseURI string
}

// GetMgmtAuth loads an Azure authorizer using a credentials file for the management
// layer
func (va AzureKeyVaultFileAuth) GetMgmtAuth() (*autorest.Authorizer, error) {
	at, err := auth.NewAuthorizerFromFile(va.BaseURI)
	return &at, err
}

// GetAuth loads an Azure authorizer using a credentials file for the Key Vault
// component specifically
func (va AzureKeyVaultFileAuth) GetAuth() (*autorest.Authorizer, error) {
	at, err := kvauth.NewAuthorizerFromFile(va.BaseURI)
	return &at, err
}
