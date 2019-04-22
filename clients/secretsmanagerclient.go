package clients

// SecretsManagerClient is the common interface used for
// interacting with any kind of backend Secrets manager.
// All integrations with a secrets manager should
// implement this interface
type SecretsManagerClient interface {
	GetSecrets(namespaces ...string) ([]*Secret, error)
}

// Secret is a common struct designed as an intermediary
// struct between a backend secrets manager, and k8s
type Secret struct {
	Name      string
	Namespace string
	Data      map[string]string
}
