package clients

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/hashicorp/vault/api"
)

// hashicorpVaultClient holds the required client and paths for integration with Hashicorp Vault
type hashicorpVaultClient struct {
	Client       *api.Client
	dataPath     string
	metadataPath string
}

// NewHashicorpVaultClient provides a new SecretsManagerClient for using Hashicorp Vault
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
	dataPath, metadataPath := setupVaultPaths(version, mount, path)
	return &hashicorpVaultClient{
		Client:       client,
		dataPath:     dataPath,
		metadataPath: metadataPath,
	}, nil
}

// setValutPaths provides the data and metadata paths for vault integration
func setupVaultPaths(version int, mount, path string) (string, string) {
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
	return dataPath, metadataPath
}

// GetSecrets will provide a slice of Secret type responses, for remote secrets located in Hashicorp Vault
func (client hashicorpVaultClient) GetSecrets(namespaces ...string) ([]*Secret, error) {
	nc := make(chan *Secret)
	wg1 := &sync.WaitGroup{}

	for _, namespace := range namespaces {
		wg1.Add(1)
		go listVaultSecrets(nc, wg1, client.Client, client.metadataPath, namespace)
	}

	go func() {
		wg1.Wait()
		close(nc)
	}()

	sc := make(chan *Secret)
	wg2 := &sync.WaitGroup{}
	for secret := range nc {
		wg2.Add(1)
		go buildVaultSecret(sc, wg2, client.Client, client.dataPath, secret.Namespace, secret.Name)
	}

	go func() {
		wg2.Wait()
		close(sc)
	}()

	secrets := make([]*Secret, 0)
	for secret := range sc {
		secrets = append(secrets, secret)
	}
	return secrets, nil
}

// listVaultSecrets will retrieve a list of secrets from Hashicorp Vault on the provided paths
func listVaultSecrets(c chan<- *Secret, wg *sync.WaitGroup, client *api.Client, metadataPath, namespace string) {
	defer wg.Done()
	secretsList, err := client.Logical().List(fmt.Sprintf("%s/%s", metadataPath, namespace))
	if err != nil {
		log.Println(err.Error())
		return
	}
	if secretsList == nil {
		return
	}
	loadVaultSecretsAtPath(c, namespace, secretsList.Data)
}

// loadVaultSecretsAtPath loads secrets from a vault response, filtering out further downstream paths
// if any are provided
func loadVaultSecretsAtPath(c chan<- *Secret, namespace string, data map[string]interface{}) {
	for k, v := range data {
		if k == "keys" && v != nil {
			for _, kv := range v.([]interface{}) {
				if kvstr, ok := kv.(string); ok && kvstr[len(kvstr)-1:] != "/" {
					c <- &Secret{Name: kvstr, Namespace: namespace}
				}
			}
		}
	}
}

// buildVaultSecret will build a Secret from secret data retrieved from the vault
func buildVaultSecret(c chan<- *Secret, wg *sync.WaitGroup, client *api.Client, dataPath, namespace, name string) {
	defer wg.Done()
	vaultSecret, err := client.Logical().Read(fmt.Sprintf("%s/%s/%s", dataPath, namespace, name))
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
	c <- &Secret{Name: name, Namespace: namespace, Data: secretData}
}

// getVersion will get the kv engine version used by the requested mount point
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
