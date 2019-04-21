package clients

type SecretsManagerClient interface {
	GetSecrets(namespaces ...string) ([]*Secret, error)
}

type Secret struct {
	Name      string
	Namespace string
	Data      map[string]string
}
