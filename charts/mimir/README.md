# mimir Helm Chart

A chart for deploying mimir onto the cluster as either a cronjob, or a deployment with a webhook (or both)

## Supported Backends

Currently both Hashicorp Vault and AWS Secrets Manager are supported

## Remote Managed Secrets

### Hashicorp Vault

For Hashicorp Vault, secrets to be managed are based on the path inside the vault when using a kv engine type. Essentially, mimir will scan the vault based on the `mount` provided to the CLI. It will, within there check for any directories matching a namespace in the cluster, and load in any secrets within that directory.

For example, a secret at the following path: `https://myvault.mydomain/v1/secret/default/example` in vault would be loaded in to the `default` namespace with the name `example`. Both kv engine v1 & v2 are supported. The tool is not limited to top level, and an optional path variable can be provided to set a root path for secrets within a mount.

### AWS Secrets Manager

*Tags only need to be added when using mimir with sync mode. They are not required when using the webhook*

Secrets managed in AWS are based on tags. Create secrets in AWS as normal, but to sync them, the following two tags should be added:

* Key: `mimir-managed`, Value: `true/false` - Sets a true or false string on if the secret should be synced with kubernetes
* Key: `mimir-paths`, Value: `{namespace1}/{secret}+{namespace2}/{secret}` - Provided list of `+` separated paths on where the secret should sync to in k8s. Path format is namespace / secret, and will be loaded into the cluster this way.

## Values for deployment

| Paramter                          | Description                                                                        | Default                   | Required                          |
| --------------------------------- | ---------------------------------------------------------------------------------- | ------------------------- | --------------------------------- |
| `serviceAccount`                  | The name of the service account to use for mimir                                   | `mimir-service-account`   | yes                               |
| `job.enabled`                     | Should the cronjob be deployed to the cluster                                      | `false`                   | yes                               |
| `job.schedule`                    | The cron schedule to run the sync                                                  | `*/5 * * * *`             | yes - If job enabled              |
| `job.restartPolicy`               | Should the cronjob pods try to restart on failure                                  | `false`                   | yes - If job enabled              |
| `image.respository`               | The repository of the mimir image                                                  | `marmotherder/mimir`      | yes                               |
| `image.tag`                       | The image tag                                                                      | `latest`                  | yes                               |
| `image.pullPolicy`                | Pull policy on the image every run                                                 | `IfNotPresent`            | yes                               |
| `hashicorpVault.enabled`          | Run sync with Hashicorp Vault                                                      | `false`                   | yes                               |
| `hashicorpVault.auth`             | Authentication to use with Hasicorp Vault - Options are: `k8s`, `approle`, `token` | `k8s`                     | yes - If vault enabled            |
| `hashicorpVault.url`              | The URL to the Hashicorp Vault                                                     | `http://vault-vault:8200` | yes - if vault is enabled         |
| `hashicorpVault.mount`            | The secrets mount in the vault                                                     | `secret`                  | yes - If vault enabled            |
| `hashicorpVault.path`             | Drilldown path in the mount to a secrets holding directory                         | `secret`                  | no                                |
| `hashicorpVault.role`             | Vault role to bind a kubernetes token to                                           | `reader`                  | yes - if auth is `k8s`            |
| `hashicorpVault.roleid`           | Approle role_id to use to authenticate with vault                                  | na                        | yes - if auth is `approle`        |
| `hashicorpVault.secretid`         | Approle secret_id to use to authenticate with vault                                | na                        | yes - if auth is `approle`        |
| `hashicorpVault.token`            | Valid vault token to authenticate with vault                                       | na                        | yes - if auth is `token`          |
| `aws.enabled`                     | Run sync with AWS Secrets manager                                                  | `false`                   | yes                               |
| `aws.auth`                        | Authentication to use with AWS - Options are `iam`, `static`, `env`, `shared`      | `iam`                     | yes - if aws enabled              |
| `aws.region`                      | The AWS region to connect to                                                       | `eu-west-1`               | yes - if aws enabled              |
| `aws.accesskey`                   | The AWS ACCESS_KEY_ID to use to authenticate with AWS                              | na                        | yes - if auth is `static`         |
| `aws.secretkey`                   | The AWS SECRET_ACCESS_KEY to use to authenticate with AWS                          | na                        | yes - if auth is `static`         |
| `aws.path`                        | The path to an AWS shared credentials file                                         | na                        | no - optional if auth is `shared` |
| `aws.profile`                     | The AWS profile to use                                                             | na                        | no - optional if auth is `shared` |
| `azure.enabled`                   | Run sync with Azure Key Vault secrets                                              | `false`                   | yes                               |
| `azure.auth`                      | Authentication to use with Azure - Options are `env` or `file`                     | `env`                     | yes                               |
| `azure.subscriptionID`            | Azure subscription ID to use. Uses `AZURE_SUBSCRIPTION_ID` env variable if not set | na                        | no                                |
| `azure.credentialsFilePath`       | The path to an Azure credentials file                                              | na                        | yes - if auth is `file`           |
| `webhook.enabled`                 | Should mimir be deployed as a webhook server in the cluster                        | `false`                   | yes                               |
| `webhook.failurePolicy`           | The k8s webhook policy to use, `Fail` or `Ignore` are supported                    | `Ignore`                  | yes - if webhook enabled          |
| `webhook.initImage.repository`    | The repository of the mimir init image                                             | `marmotherder/mimir-init` | yes - if webhook enabled          |
| `webhook.initImage.tag`           | The image tag                                                                      | `latest`                  | yes - if webhook enabled          |
| `webhook.initImage.pullPolicy`    | Pull policy on the image for hooks                                                 | `IfNotPresent`            | yes - if webhook enabled          |
