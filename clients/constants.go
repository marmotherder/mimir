package clients

// SecretsManager denotes what secrets manager backend
// was used for a particuar secret
type SecretsManager string

const (
	// HashicorpVault denotes the secret was managed
	// by Hashicorp Vault
	HashicorpVault SecretsManager = "hashicorp-vault"
	// AWS denotes the secret was managed by AWS
	// secrets manager
	AWS SecretsManager = "aws"
	// Azure denotes the secret was managed by Azure
	// Key Vault
	// TODO - Implement Azure Key Vault solution
	Azure SecretsManager = "azure"
	// GCP denotes the secret was managed by GCP
	// TODO - Implement a secrets management
	// solution for GCP, KMS?
	GCP SecretsManager = "gcp"
	// Managed is the common tag/annotation denoting
	// that the secret is managed by mimir
	Managed string = "mimir-managed"
	// Paths is the common tag to use to speify what
	// paths to load the secret into k8s under
	Paths string = "mimir-paths"
	// Source is the common annotation to denote
	// where the secret was sourced from in k8s
	Source string = "mimir-source"
	// Hook is a reference string per server that
	// allows multiple hooks to co-exist in the
	// same cluster
	Hook string = "mimir-hook"
	// Remote is the path/name of the remote secret
	Remote string = "mimir-remote"
	// Local is an override. When set, the secret will
	// be created with the name given to this attribute,
	// rather than the pod name
	Local string = "mimir-local"
	// Path is the local container path the secrets
	// should be mounted to
	Path string = "mimir-path"
	// Env is a switch that when set, makes mimir
	// patch the pod to inject all the keys of the
	// secret to the containers as environment vars
	Env string = "mimir-env"
)
