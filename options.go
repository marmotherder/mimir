package main

// Options is the common mimir options available
type Options struct {
	Backend        string `short:"b" long:"backend" choice:"hashicorpvault" choice:"aws" description:"The secrets manager backend to be used" required:"true"`
	IsPod          bool   `short:"i" long:"ispod" description:"Is the application being run within a pod?"`
	KubeconfigPath string `short:"k" long:"kcpath" description:"An absolute path to a valid kube config file"`
}

// HashiCorpVaultOptions is the base configuration options for Hashicorp Valut
type HashiCorpVaultOptions struct {
	Authentication string `short:"a" long:"auth" choice:"k8s" choice:"approle" choice:"token" description:"Authentication method to use with Hashicorp Vault" required:"true"`
	URL            string `short:"u" long:"url" description:"The base URL to the Hashicorp Vault instance" required:"true"`
	Mount          string `short:"m" long:"mount" description:"Which mount to attach to in the vault" required:"true"`
	Path           string `short:"p" long:"path" description:"Optional to provide a root path within the mount on where to look for secrets"`
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

// HashiCorpVaultOptions is the base configuration options for AWS Secrets Manager
type AWSOptions struct {
	Authentication string `short:"a" long:"auth" choice:"iam" choice:"static" choice:"env" choice:"shared" description:"Authentication method to use with AWS" required:"true"`
	Region         string `short:"r" long:"region" description:"The AWS region to connect to" required:"true"`
}

// HashicorpVaultTokenOptions allows providing the AWS ACCESS_KEY_ID and the AWS SECRET_ACCESS_KEY to AWS Secrets Manager via the CLI
type AWSCredentialsOptions struct {
	AccessKeyID     string `short:"e" long:"accesskey" description:"The AWS ACCESS_KEY_ID variable to use" required:"true"`
	SecretAccessKey string `short:"s" long:"secretkey" description:"The AWS SECRET_ACCESS_KEY variable to use" required:"true"`
}

// AWSSharedOptions allows providing the file path and profile to use for shared credentials via the CLI
type AWSSharedOptions struct {
	Path    string `short:"p" long:"path" description:"The absolute path to the AWS credentials file"`
	Profile string `short:"f" long:"profile" description:"The AWS profile to use"`
}
