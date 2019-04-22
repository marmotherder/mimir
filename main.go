package main

import (
	"log"
	"os"

	"mimir/clients"

	"github.com/jessevdk/go-flags"
)

func main() {
	var opts Options
	parseArgs(&opts)

	switch opts.Backend {
	case "hashicorpvault":
		var hvOpts HashiCorpVaultOptions
		parseArgs(&hvOpts)
		switch hvOpts.Authentication {
		case "k8s":
			var hvK8SOpts HashicorpVaultK8SOptions
			parseArgs(&hvK8SOpts)
			runHashiCorpVault(opts, hvOpts, clients.HashicorpVaultK8SAuth{IsPod: opts.IsPod, Role: hvK8SOpts.Role, ConfigPath: opts.KubeconfigPath})
		case "approle":
			var hvAppRoleOpts HashicorpVaultAppRoleOptions
			parseArgs(&hvAppRoleOpts)
			runHashiCorpVault(opts, hvOpts, clients.HashicorpVaultApproleAuth{RoleID: hvAppRoleOpts.RoleID, SecretID: hvAppRoleOpts.SecretID})
		case "token":
			var hvTokenOpts HashicorpVaultTokenOptions
			parseArgs(&hvTokenOpts)
			runHashiCorpVault(opts, hvOpts, clients.HashicorpVaultTokenAuth{Token: hvTokenOpts.Token})
		default:
			log.Fatal("Unknown Hashicorp Vault authentication type\n")
		}
	case "aws":
		var awsOpts AWSOptions
		parseArgs(&awsOpts)
		switch awsOpts.Authentication {
		case "iam":
			runAWS(opts, awsOpts, &clients.AWSIAMAuth{})
		case "static":
			var staticAWSOpts AWSCredentialsOptions
			parseArgs(&staticAWSOpts)
			runAWS(opts, awsOpts, &clients.AWSStaticCredentialsAuth{AccessKeyID: staticAWSOpts.AccessKeyID, SecretAccessKey: staticAWSOpts.SecretAccessKey})
		case "env":
			runAWS(opts, awsOpts, &clients.AWSEnvironmentAuth{})
		case "shared":
			var awsSharedOpts AWSSharedOptions
			parseArgs(&awsSharedOpts)
			runAWS(opts, awsOpts, &clients.AWSSharedCredentialsAuth{Path: awsSharedOpts.Path, Profile: awsSharedOpts.Profile})
		default:
			log.Fatal("Unknown AWS authentication type\n")
		}
	default:
		log.Fatal("Failed to load a configured secrets backend properly\n")
	}
}

// parseArgs parses the cli flags, allowing a common point to parse later downstream options when
// building config from multiple structs
func parseArgs(opts interface{}) {
	parser := flags.NewParser(opts, flags.IgnoreUnknown)
	_, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// runHashiCorpVault is the entrypoint for a mimir run using Hashicorp Vault
func runHashiCorpVault(opts Options, hvOpts HashiCorpVaultOptions, auth clients.HashicorpVaultAuth) {
	client, err := clients.NewHashicorpVaultClient(hvOpts.Path, hvOpts.URL, hvOpts.Mount, auth)
	if err != nil {
		log.Fatalln(err.Error())
	}
	run(opts, client, clients.HashicorpVault)
}

// runAWS is the entrypoint for a mimir run using AWS Secrets Manager
func runAWS(opts Options, awsOpts AWSOptions, auth clients.AWSSecretsAuth) {
	auth.SetRegion(awsOpts.Region)
	client, err := clients.NewAWSSecretsClient(auth)
	if err != nil {
		log.Fatalln(err.Error())
	}
	run(opts, client, clients.AWS)
}

// run performs a run of mimir secret syncing for the given backend
func run(opts Options, smc clients.SecretsManagerClient, mgr clients.SecretsManager) {
	kc, err := clients.NewK8SClient(opts.IsPod, opts.KubeconfigPath)
	if err != nil {
		log.Fatalln(err.Error())
	}
	namespaces, err := clients.GetNamespaces(kc)
	if err != nil {
		log.Fatalln(err.Error())
	}
	secrets, err := smc.GetSecrets(namespaces...)
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = clients.ManageSecrets(kc, mgr, secrets...)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
