# mimir

[![Build Status](https://travis-ci.org/marmotherder/mimir.svg?branch=master)](https://travis-ci.org/marmotherder/mimir)

A commandline based tool for synchronising secrets between kubernetes, and an external hosted secrets management platform

## Supported Backends

Currently both Hashicorp Vault and AWS Secrets Manager are supported

## Remote Managed Secrets

### Hashicorp Vault

For Hashicorp Vault, secrets to be managed are based on the path inside the vault when using a kv engine type. Essentially, mimir will scan the vault based on the `mount` provided to the CLI. It will, within there check for any directories matching a namespace in the cluster, and load in any secrets within that directory.

For example, a secret at the following path: `https://myvault.mydomain/v1/secret/default/example` in vault would be loaded in to the `default` namespace with the name `example`. Both kv engine v1 & v2 are supported. The tool is not limited to top level, and an optional path variable can be provided to set a root path for secrets within a mount.

### AWS Secrets Manager

Secrets managed in AWS are based on tags. Create secrets in AWS as normal, but to sync them, the following two tags should be added:

* Key: `mimir-managed`, Value: `true/false` - Sets a true or false string on if the secret should be synced with kubernetes
* Key: `mimir-paths`, Value: `{namespace1}/{secret}+{namespace2}/{secret}` - Provided list of `+` separated paths on where the secret should sync to in k8s. Path format is namespace / secret, and will be loaded into the cluster this way.

## Running mimir

Mimir can be run via commandline on any windows/macOS/linux system via a command line interface. Alternatively, the provided helm chart at that `chart` path will allow you to deploy the application onto a cluster, where it will run as a crojob.

For running it via the backend, the following are the top level CLI arguments that must be passed in.

| Long      | Short | Description                                  | Choices                 | Required                          |
| --------- | ----- | -------------------------------------------- | ----------------------- | --------------------------------- |
| `backend` | `b`   | The secrets manager backend to be used       | `hashicorpvault`, `aws` | yes                               |
| `ispod`   | `i`   | Is the application being run within a pod?   |                         | no - Defaults to false if not set |
| `kcpath`  | `k`   | An absolute path to a valid kube config file |                         | no - Takes from home if not set   |

### Running for Hashicorp Vault

| Long       | Short | Description                                                                   | Choices                   | Required                   |
| ---------- | ----- | ----------------------------------------------------------------------------- | ------------------------- | -------------------------- |
| `auth`     | `a`   | Authentication method to use with Hashicorp Vault                             | `k8s`, `approle`, `token` | yes                        |
| `url`      | `u`   | The base URL to the Hashicorp Vault instance                                  |                           | yes                        |
| `mount`    | `m`   | Which mount to attach to in the vault                                         |                           | yes                        |
| `path`     | `p`   | Optional to provide a root path within the mount on where to look for secrets |                           | no                         |
| `role`     | `r`   | The Hashicorp Vault role to bind the K8S token against                        |                           | yes - if auth is `k8s`     |
| `roleid`   | `r`   | The Hashicorp Vault role ID                                                   |                           | yes - if auth is `approle` |
| `secretid` | `s`   | The Hashicorp Vault secret ID                                                 |                           | yes - if auth is `approle` |
| `token`    | `t`   | The Hashicorp Vault token                                                     |                           | yes - if auth is `token`   |

### Running for AWS SecretsManager

| Long        | Short | Description                                   | Choices                          | Required                  |
| ----------- | ----- | --------------------------------------------- | -------------------------------- | ------------------------- |
| `auth`      | `a`   | Authentication method to use with AWS         | `iam`, `static`, `env`, `shared` | yes                       |
| `region`    | `r`   | The AWS region to connect to                  |                                  | yes                       |
| `accesskey` | `e`   | The AWS ACCESS_KEY_ID variable to use         |                                  | yes - if auth is `static` |
| `secretkey` | `s`   | The AWS SECRET_ACCESS_KEY variable to use     |                                  | yes - if auth is `static` |
| `path`      | `p`   | The absolute path to the AWS credentials file |                                  | no                        |
| `profile`   | `f`   | The AWS profile to use                        |                                  | no                        |