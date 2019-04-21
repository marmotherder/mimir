package clients

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getHomeDir() string {
	if runtime.GOOS == "windows" {
		if userHome := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH"); userHome != "" {
			return userHome
		}
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}

func getRestConfig(isPod bool, configPath ...string) (*rest.Config, error) {
	if len(configPath) > 0 {
		return clientcmd.BuildConfigFromFlags("", configPath[0])
	}
	if isPod {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/.kube/config", getHomeDir()))
}

func NewK8SClient(isPod bool, configPath ...string) (*kubernetes.Clientset, error) {
	config, err := getRestConfig(isPod, configPath...)

	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("Failed to load kubernetes config")
	}
	return kubernetes.NewForConfig(config)
}

func GetNamespaces(client *kubernetes.Clientset) ([]string, error) {
	k8sNamespaces, err := client.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	namespaces := make([]string, 0)
	for _, k8sNamespace := range k8sNamespaces.Items {
		namespaces = append(namespaces, k8sNamespace.Name)
	}
	return namespaces, nil
}

func ManageSecrets(client *kubernetes.Clientset, mgr SecretsManager, secrets ...*Secret) error {
	namespaces, err := GetNamespaces(client)
	if err != nil {
		return err
	}
	for _, namespace := range namespaces {
		nsSecrets := make([]*Secret, 0)
		func() {
			for _, secret := range secrets {
				if secret.Namespace == namespace {
					nsSecrets = append(nsSecrets, secret)
				}
			}
		}()

		k8sSecrets, err := client.CoreV1().Secrets(namespace).List(meta_v1.ListOptions{})
		if err != nil {
			return err
		}

		managedSecrets := getManagedSecrets(k8sSecrets.Items, mgr)

		for _, nsSecret := range nsSecrets {
			hasSecret := func() bool {
				for _, k8sSecret := range managedSecrets {
					if nsSecret.Name == k8sSecret.Name {
						return true
					}
				}
				return false
			}()
			if hasSecret {
				if _, err := client.CoreV1().Secrets(namespace).Update(buildK8sSecret(nsSecret, mgr)); err != nil {
					return err
				}
				log.Printf("Updated secret: %s in namespace %s\n", nsSecret.Name, nsSecret.Namespace)
			} else {
				if _, err := client.CoreV1().Secrets(namespace).Create(buildK8sSecret(nsSecret, mgr)); err != nil {
					return err
				}
				log.Printf("Created secret: %s in namespace %s\n", nsSecret.Name, nsSecret.Namespace)
			}
		}

		for _, k8sSecret := range managedSecrets {
			nsSecret := func() *Secret {
				for _, nsSecret := range nsSecrets {
					if nsSecret.Name == k8sSecret.Name {
						return nsSecret
					}
				}
				return nil
			}()
			if nsSecret == nil {
				if err := client.CoreV1().Secrets(namespace).Delete(k8sSecret.Name, &meta_v1.DeleteOptions{}); err != nil {
					return nil
				}
				log.Printf("Deleted secret: %s in namespace %s\n", k8sSecret.Name, namespace)
			}
		}
	}
	return nil
}

func getManagedSecrets(secrets []core_v1.Secret, mgr SecretsManager) []core_v1.Secret {
	managedSecrets := make([]core_v1.Secret, 0)
	for _, secret := range secrets {
		isManaged := false
		isSource := false
		for k, v := range secret.Annotations {
			if k == Managed && v == "true" {
				isManaged = true
			}
			if k == Source && v == string(mgr) {
				isSource = true
			}
		}
		if isManaged && isSource {
			managedSecrets = append(managedSecrets, secret)
		}
	}
	return managedSecrets
}

func buildK8sSecret(secret *Secret, mgr SecretsManager) *core_v1.Secret {
	data := make(map[string][]byte)
	for k, v := range secret.Data {
		data[k] = []byte(v)
	}
	return &core_v1.Secret{
		Type: core_v1.SecretTypeOpaque,
		Data: data,
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Annotations: map[string]string{
				Managed: "true",
				Source:  string(mgr),
			},
		},
	}
}
