package clients

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/vault/api"
)

type hashicorpVaultClient struct {
	Client       *api.Client
	dataPath     string
	metadataPath string
}

func NewHashicorpVaultClient(path, url, mount string, auth HashicorpVaultAuth) (SecretsManagerClient, error) {
	client, err := api.NewClient(&api.Config{
		Address: url,
	})
	if err != nil {
		return nil, err
	}
	err = auth.GetToken(client)
	if err != nil {
		return nil, err
	}
	version, err := getVersion(mount, client)
	if err != nil {
		return nil, err
	}
	dataPath := mount
	metadataPath := mount
	if path != "" {
		dataPath = fmt.Sprintf("%s/%s", mount, path)
		metadataPath = fmt.Sprintf("%s/%s", mount, path)
	}
	if version > 1 {
		if path != "" {
			dataPath = fmt.Sprintf("%s/data/%s", mount, path)
			metadataPath = fmt.Sprintf("%s/metadata/%s", mount, path)
		} else {
			dataPath = fmt.Sprintf("%s/data", mount)
			metadataPath = fmt.Sprintf("%s/metadata", mount)
		}
	}
	return &hashicorpVaultClient{
		Client:       client,
		dataPath:     dataPath,
		metadataPath: metadataPath,
	}, nil
}

func (client hashicorpVaultClient) GetSecrets(namespaces ...string) ([]*Secret, error) {
	secrets := make([]*Secret, 0)
	for _, namespace := range namespaces {
		vaultSecrets, err := client.listSecrets(namespace)
		if err != nil {
			log.Println(err.Error())
		}
		if vaultSecrets != nil {
			for _, secretName := range vaultSecrets {
				vaultSecret, err := client.Client.Logical().Read(fmt.Sprintf("%s/%s/%s", client.dataPath, namespace, secretName))
				if err != nil {
					log.Println(err.Error())
				}
				secretData := make(map[string]string)
				if data, ok := vaultSecret.Data["data"].(map[string]interface{}); ok {
					for k, v := range data {
						if vstr, ok := v.(string); ok {
							secretData[k] = vstr
						}
					}
				}
				secrets = append(secrets, &Secret{Name: secretName, Namespace: namespace, Data: secretData})
			}
		}
	}
	return secrets, nil
}

func (client hashicorpVaultClient) listSecrets(namespace string) ([]string, error) {
	secretsList, err := client.Client.Logical().List(fmt.Sprintf("%s/%s", client.metadataPath, namespace))
	if err != nil {
		return nil, err
	}
	if secretsList == nil {
		return nil, nil
	}

	secrets := make([]string, 0)

	for k, v := range secretsList.Data {
		if k == "keys" && v != nil {
			for _, kv := range v.([]interface{}) {
				if kvstr, ok := kv.(string); ok && kvstr[len(kvstr)-1:] != "/" {
					secrets = append(secrets, kvstr)
				}
			}
		}
	}

	return secrets, nil
}

func getVersion(mount string, client *api.Client) (int, error) {
	mounts, err := client.Sys().ListMounts()
	if err != nil {
		return 0, err
	}
	for mountKey, mountConfig := range mounts {
		if mountKey == fmt.Sprintf("%s/", mount) {
			if mountConfig.Type != "kv" {
				return 0, errors.New("Only kv engine types are supported for sync")
			}
			version, err := strconv.Atoi(mountConfig.Options["version"])
			if err != nil {
				return 0, err
			}
			if version <= 0 {
				return 0, errors.New("Could not locate the version of the kv engine")
			}
			return version, nil
		}
	}
	return 0, errors.New("Did not find the requested mount")
}
