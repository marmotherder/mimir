package clients

import (
	"errors"
	"io/ioutil"

	"github.com/hashicorp/vault/api"
)

// HashicorpVaultAuth interface provides a common function set to authenticate Hashicorp Vault
type HashicorpVaultAuth interface {
	GetToken(client *api.Client) error
}

// HashicorpVaultK8SAuth contains auth information for using kubernetes authentication method
type HashicorpVaultK8SAuth struct {
	IsPod      bool
	Role       string
	ConfigPath string
}

// GetToken retrieves a valid Hashicorp Vault token via kubernetes authentication method for integrating with the vault
func (auth HashicorpVaultK8SAuth) GetToken(client *api.Client) error {
	if auth.Role == "" {
		return errors.New("No valid vault role provided")
	}
	token := ""
	if auth.IsPod {
		tokenBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
		if err == nil {
			token = string(tokenBytes)
		}
	}
	if token == "" {
		config, err := getRestConfig(auth.IsPod, auth.ConfigPath)
		if err != nil {
			return err
		}
		token = config.BearerToken
	}
	if token == "" {
		return errors.New("Failed to load kubernetes client token")
	}

	vaultToken, err := client.Logical().Write("auth/kubernetes/login", map[string]interface{}{"jwt": token, "role": auth.Role})
	if err != nil {
		return err
	}
	client.SetToken(vaultToken.Auth.ClientToken)
	return nil
}

// HashicorpVaultTokenAuth contains auth information for using a pre-provided token to authenticate
type HashicorpVaultTokenAuth struct {
	Token string
}

// GetToken retrieves the pre-provided token
func (auth HashicorpVaultTokenAuth) GetToken(client *api.Client) error {
	if auth.Token == "" {
		return errors.New("No valid vault token set")
	}
	client.SetToken(auth.Token)
	return nil
}

// HashicorpVaultApproleAuth contains auth information for using the approle auth method
type HashicorpVaultApproleAuth struct {
	RoleID   string
	SecretID string
}

// GetToken retrieves a valid Hashicorp Vault token via approle authentication method for integrating with the vault
func (auth HashicorpVaultApproleAuth) GetToken(client *api.Client) error {
	if auth.RoleID == "" {
		return errors.New("No valid role id set")
	}
	if auth.SecretID == "" {
		errors.New("No valid secret id set")
	}
	vaultToken, err := client.Logical().Write("auth/approle/login", map[string]interface{}{"role_id": auth.RoleID, "secret_id": auth.SecretID})
	if err != nil {
		return err
	}
	client.SetToken(vaultToken.Auth.ClientToken)
	return nil
}
