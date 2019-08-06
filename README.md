# mimir

[![Build Status](https://travis-ci.org/marmotherder/mimir.svg?branch=master)](https://travis-ci.org/marmotherder/mimir)
[![Go Report Card](https://goreportcard.com/badge/github.com/marmotherder/mimir)](https://goreportcard.com/report/github.com/marmotherder/mimir)

A commandline based tool for synchronising secrets between kubernetes, and an external hosted secrets management platform

Also supports running as a web server for listening to requests as a [Kubernetes Admissions Controller](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)

## Supported Backends

Currently both Hashicorp Vault and AWS Secrets Manager are supported

## Running as a Admission Controller

mimir supports a deployment of itself onto a k8s cluster to act as an Admission Controller in the cluster. When in this setup, mimir will deploy a webhook and itself, and then act as a hook for all pod creation and deletion requests. For pods that have mimir annotations, the hook will attempt to create a secret sourced from a remote secrets manager, and patch the pod to load these secrets. At delete it will try to delete the secret to clean up.

### mimir pod annotations

The following annotations are suppored at pod level. Using these annotations in a pod spec will trigger a mutation via the mimir Admission controller

* `mimir-hook` - The name of the webhook that should be used to lookup the secrets. This value will be set at the deploy time of mimir, and then referenced in the pod
* `mimir-remote` - The name/path of the secret in the remote secret manager. For AWS Secret Manager, this should just be the name of the secret, for Hashicorp Vault, it should be the path (relative to the mount provided to mimir server)
* `mimir-path` - The path on all containers in the pod that the remote secret should be mounted to as files (optional)
* `mimir-env` - A switch, which when set as "true", will load all the keys in the secret as an environment variable in all the containers in the pod (optional)
* `mimir-local` - Overrides the name of the generated secret with what is provided here (optional)

## Remote Managed Secrets

### Hashicorp Vault

For Hashicorp Vault, secrets to be managed are based on the path inside the vault when using a kv engine type. Essentially, mimir will scan the vault based on the `mount` provided to the CLI. It will, within there check for any directories matching a namespace in the cluster, and load in any secrets within that directory.

For example, a secret at the following path: `https://myvault.mydomain/v1/secret/default/example` in vault would be loaded in to the `default` namespace with the name `example`. Both kv engine v1 & v2 are supported. The tool is not limited to top level, and an optional path variable can be provided to set a root path for secrets within a mount.

### AWS Secrets Manager

Secrets managed in AWS are based on tags. Create secrets in AWS as normal, but to sync them, the following two tags should be added:

* Key: `mimir-managed`, Value: `true/false` - Sets a true or false string on if the secret should be synced with kubernetes
* Key: `mimir-paths`, Value: `{namespace1}/{secret}+{namespace2}/{secret}` - Provided list of `+` separated paths on where the secret should sync to in k8s. Path format is namespace / secret, and will be loaded into the cluster this way.

## Running mimir

Mimir can be run via commandline on any windows/macOS/linux system via a command line interface. Alternatively, the provided helm charts at that `charts` path will allow you to deploy the application onto a cluster.
The chart `mimir-service-account` should be deployed before the basic `mimir` chart.

For running it via the backend, the following are the top level CLI arguments that must be passed in.

| Long      | Short | Description                                                    | Choices                 | Required                          |
| --------- | ----- | -------------------------------------------------------------- | ----------------------- | --------------------------------- |
| `backend` | `b`   | The secrets manager backend to be used                         | `hashicorpvault`, `aws` | yes                               |
| `ispod`   | `i`   | Is the application being run within a pod?                     |                         | no - Defaults to false if not set |
| `kcpath`  | `k`   | An absolute path to a valid kube config file                   |                         | no - Takes from home if not set   |
| `server`  | `o`   | Should mimir run as a webserver for listening to k8s webhooks? |                         | no - Defaults to false if not set |

### Running as a webhook server

*Running mimir as a server for webhooks will not alone build/deploy the k8s configuration for listening to webhooks. As such it is highly recommended you do not run these options outside of the provided helm deployment*

| Long      | Short | Description                                                                       | Default   | Required  |
| --------- | ----- | --------------------------------------------------------------------------------- | --------- | --------- |
| `port`    | `d`   | The port for the server to listen on                                              | `443`     | yes       |
| `cert`    | `c`   | Filesystem path to a PEM encoded CA signed certificate                            |           | yes       |
| `key`     | `l`   | Filesystem path to a PEM encoded private key for a tls cert                       |           | yes       |
| `hook`    | `h`   | Reference sting for the hook. Allows multiple hooks to run in the same cluster    |           | yes       |

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
