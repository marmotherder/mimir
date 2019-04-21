package clients

type SecretsManager string

const (
	HashicorpVault SecretsManager = "hashicorp-vault"
	AWS            SecretsManager = "aws"
	Azure          SecretsManager = "azure"
	GCP            SecretsManager = "gcp"
	Managed        string         = "mimir-managed"
	Paths          string         = "mimir-paths"
	Source         string         = "mimir-source"
)
