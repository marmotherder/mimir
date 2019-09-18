package main

import (
	"errors"

	"github.com/marmotherder/mimir/clients"
)

// loadClient is a helper function to load a live client for a configured secrets manager backend
func loadClient() (clients.SecretsManagerClient, clients.SecretsManager, error) {
	switch opts.Backend {
	case "hashicorpvault":
		var hvOpts HashiCorpVaultOptions
		parseArgs(&hvOpts)
		switch hvOpts.Authentication {
		case "k8s":
			var hvK8SOpts HashicorpVaultK8SOptions
			parseArgs(&hvK8SOpts)
			smc, mgr := loadHashiCorpVaultClient(opts, hvOpts, clients.HashicorpVaultK8SAuth{IsPod: opts.IsPod, Role: hvK8SOpts.Role, ConfigPath: opts.KubeconfigPath})
			return smc, mgr, nil
		case "approle":
			var hvAppRoleOpts HashicorpVaultAppRoleOptions
			parseArgs(&hvAppRoleOpts)
			smc, mgr := loadHashiCorpVaultClient(opts, hvOpts, clients.HashicorpVaultApproleAuth{RoleID: hvAppRoleOpts.RoleID, SecretID: hvAppRoleOpts.SecretID})
			return smc, mgr, nil
		case "token":
			var hvTokenOpts HashicorpVaultTokenOptions
			parseArgs(&hvTokenOpts)
			smc, mgr := loadHashiCorpVaultClient(opts, hvOpts, clients.HashicorpVaultTokenAuth{Token: hvTokenOpts.Token})
			return smc, mgr, nil
		default:
			return nil, "", errors.New("Unknown Hashicorp Vault authentication type")
		}
	case "aws":
		var awsOpts AWSOptions
		parseArgs(&awsOpts)
		switch awsOpts.Authentication {
		case "iam":
			smc, mgr := loadAWSClient(opts, awsOpts, &clients.AWSIAMAuth{})
			return smc, mgr, nil
		case "static":
			var staticAWSOpts AWSCredentialsOptions
			parseArgs(&staticAWSOpts)
			smc, mgr := loadAWSClient(opts, awsOpts, &clients.AWSStaticCredentialsAuth{AccessKeyID: staticAWSOpts.AccessKeyID, SecretAccessKey: staticAWSOpts.SecretAccessKey})
			return smc, mgr, nil
		case "env":
			smc, mgr := loadAWSClient(opts, awsOpts, &clients.AWSEnvironmentAuth{})
			return smc, mgr, nil
		case "shared":
			var awsSharedOpts AWSSharedOptions
			parseArgs(&awsSharedOpts)
			smc, mgr := loadAWSClient(opts, awsOpts, &clients.AWSSharedCredentialsAuth{Path: awsSharedOpts.Path, Profile: awsSharedOpts.Profile})
			return smc, mgr, nil
		default:
			return nil, "", errors.New("Unknown AWS authentication type")
		}
	case "azure":
		var azOpts AzureKeyVaultOptions
		parseArgs(&azOpts)
		switch azOpts.Authentication {
		case "env":
			smc, mgr := loadAzureKeyVaultClient(opts, azOpts, &clients.AzureKeyVaultEnvironmentAuth{})
			return smc, mgr, nil
		case "file":
			var azFileOpts AzureKeyVaultFileOptions
			parseArgs(&azFileOpts)
			smc, mgr := loadAzureKeyVaultClient(opts, azOpts, &clients.AzureKeyVaultFileAuth{BaseURI: azFileOpts.FilePath})
			return smc, mgr, nil
		default:
			return nil, "", errors.New("Unknown Azure authentication type")
		}
	default:
		return nil, "", errors.New("Failed to load a configured secrets backend properly")
	}
}
