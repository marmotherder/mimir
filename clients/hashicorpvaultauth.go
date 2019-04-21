package clients

import (
	"errors"
	"io/ioutil"

	"github.com/hashicorp/vault/api"
)

type HashicorpVaultAuth interface {
	GetToken(client *api.Client) error
}

type HashicorpVaultK8SAuth struct {
	IsPod      bool
	Role       string
	ConfigPath *string
}

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
		config, err := getRestConfig(auth.IsPod, *auth.ConfigPath)
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

type HashicorpVaultTokenAuth struct {
	Token string
}

func (auth HashicorpVaultTokenAuth) GetToken(client *api.Client) error {
	if auth.Token == "" {
		return errors.New("No valid vault token set")
	}
	client.SetToken(auth.Token)
	return nil
}

type HashicorpVaultApproleAuth struct {
	RoleID   string
	SecretID string
}

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
