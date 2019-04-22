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
)
