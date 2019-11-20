package main

// Options is the common mimir options available
type Options struct {
	ServerMode     bool    `short:"o" long:"server" description:"Should the application run as a webserver?"`
	Backend        string  `short:"b" long:"backend" choice:"hashicorpvault" choice:"aws" choice:"azure" description:"The secrets manager backend to be used" required:"true"`
	IsPod          bool    `short:"i" long:"ispod" description:"Is the application being run within a pod?"`
	KubeconfigPath *string `short:"k" long:"kcpath" description:"An absolute path to a valid kube config file"`
}

// ServerOptions is used for the webhook server specific configuration
type ServerOptions struct {
	ServerPort  int    `short:"d" long:"port" description:"Port to run the server against" default:"443"`
	TLSCertPath string `short:"c" long:"cert" description:"Path to the TLS certificate" required:"true"`
	TLSKeyPath  string `short:"l" long:"key" description:"Path to the TLS key" required:"true"`
	Hook        string `short:"h" long:"hook" description:"The identifier for this webhook" required:"true"`
}

// HashiCorpVaultOptions is the base configuration options for Hashicorp Valut
type HashiCorpVaultOptions struct {
	Authentication string `short:"a" long:"auth" choice:"k8s" choice:"approle" choice:"token" description:"Authentication method to use with Hashicorp Vault" required:"true"`
	URL            string `short:"u" long:"url" description:"The base URL to the Hashicorp Vault instance" required:"true"`
	Mount          string `short:"m" long:"mount" description:"Which mount to attach to in the vault" required:"true"`
	Path           string `short:"p" long:"path" description:"Optional to provide a root path within the mount on where to look for secrets"`
	SkipTLSVerify  bool   `short:"f" long:"skip" description:"Optional flag to specify if https calls to vault should verify the TLS certificate chain"`
}

// HashicorpVaultK8SOptions allows providing the Hashicorp Vault role to bind to via the CLI
type HashicorpVaultK8SOptions struct {
	Role string `short:"r" long:"role" description:"The Hashicorp Vault role to bind the K8S token against" required:"true"`
}

// HashicorpVaultAppRoleOptions allows providing the role_id and secret_id via the CLI
type HashicorpVaultAppRoleOptions struct {
	RoleID   string `short:"r" long:"roleid" description:"The Hashicorp Vault role ID" required:"true"`
	SecretID string `short:"s" long:"secretid" description:"The Hashicorp Vault secret ID" required:"true"`
}

// HashicorpVaultTokenOptions allows providing an authentication token to Hashicorp Vault via the CLI
type HashicorpVaultTokenOptions struct {
	Token string `short:"t" long:"secretid" description:"The Hashicorp Vault token" required:"true"`
}

// AWSOptions is the base configuration options for AWS Secrets Manager
type AWSOptions struct {
	Authentication string `short:"a" long:"auth" choice:"iam" choice:"static" choice:"env" choice:"shared" description:"Authentication method to use with AWS" required:"true"`
	Region         string `short:"r" long:"region" description:"The AWS region to connect to" required:"true"`
}

// AWSCredentialsOptions allows providing the AWS ACCESS_KEY_ID and the AWS SECRET_ACCESS_KEY to AWS Secrets Manager via the CLI
type AWSCredentialsOptions struct {
	AccessKeyID     string `short:"e" long:"accesskey" description:"The AWS ACCESS_KEY_ID variable to use" required:"true"`
	SecretAccessKey string `short:"s" long:"secretkey" description:"The AWS SECRET_ACCESS_KEY variable to use" required:"true"`
}

// AWSSharedOptions allows providing the file path and profile to use for shared credentials via the CLI
type AWSSharedOptions struct {
	Path    string `short:"p" long:"path" description:"The absolute path to the AWS credentials file"`
	Profile string `short:"f" long:"profile" description:"The AWS profile to use"`
}

// AzureKeyVaultOptions is the base configuration options for Azure Key Vault
type AzureKeyVaultOptions struct {
	Authentication string `short:"a" long:"auth" choice:"env" choice:"file" description:"Authentication method to use with Azure" required:"true"`
	SubscriptionID string `short:"s" long:"subid" description:"The subscription ID to use, otherwise it takes from AZURE_SUBSCRIPTION_ID environmemnt variable"`
}

// AzureKeyVaultFileOptions provides a simple file path if using a credentials file for authentication
type AzureKeyVaultFileOptions struct {
	FilePath string `short:"f" long:"path" description:"Path to the Azure credentials file to use for authentication"`
}
