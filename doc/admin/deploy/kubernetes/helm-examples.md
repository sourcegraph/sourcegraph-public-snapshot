# Helm Chart Configuration Examples

This section outlines a few common scenarios for creating custom configurations. For an exhaustive list of configuration options please see our [Sourcegraph Helm Chart README.md in Github.](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/README.md)

- [Using external PostgreSQL databases](#using-external-postgresql-databases)
- [Using external Redis instances](#using-external-redis-instances)
- [Using external Object Storage](#using-external-object-storage)
- [Using SSH to clone repositories](#using-ssh-to-clone-repositories)

If your customization needs are not covered below please email <support@sourcegraph.com> for assistance.

## Using external PostgreSQL databases

The default Sourcecgraph deployment ships three separate Postgres instances: [codeinsights-db.StatefulSet.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/templates/codeinsights-db/codeinsights-db.StatefulSet.yaml), [codeintel-db.StatefulSet.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/templates/codeintel-db/codeintel-db.StatefulSet.yaml), and [pgsqlStatefulSet.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/templates/pgsql/pgsql.StatefulSet.yaml). All three can be disabled individually and replaced with external Postgres instances.

To use external PostgreSQL databases, first review our [general recommendations](https://docs.sourcegraph.com/admin/external_services/postgres#using-your-own-postgresql-server) and [required postgres permissions](https://docs.sourcegraph.com/admin/external_services/postgres#postgres-permissions-and-database-migrations).

> An example of this approach can be found in our Helm chart [using external databases](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/external-databases) example on Github.

We recommend storing the database credentials in [Secrets] created outside of the helm chart and managed in a secure manner. The Secrets should be deployed to the same namespace as the existing Sourcegraph deployment. Each database requires its own Secret and should follow the following format. The Secret name can be customized as desired:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pgsql-credentials
data:
  # notes: secrets data has to be base64-encoded
  database: ""
  host: "" # example: pgsql.database.example.com
  password: ""
  port: ""
  user: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: codeintel-db-credentials
data:
  # notes: secrets data has to be base64-encoded
  database: ""
  host: ""
  password: ""
  port: ""
  user: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: codeinsights-db-credentials
data:
  # notes: secrets data has to be base64-encoded
  database: ""
  host: ""
  password: ""
  port: ""
  user: ""
```

Set the Secret name your `override.yaml` by configuring the `auth.existingSecret` key for each database. A full example can be seen in this [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/examples/external-databases/override.yaml)

```yaml
codeIntelDB:
  enabled: false # disables deployment of the database
  auth:
    existingSecret: codeintel-db-credentials

codeInsightsDB:
  enabled: false
  auth:
    existingSecret: codeinsights-db-credentials

pgsql:
  enabled: false
  auth:
    existingSecret: pgsql-credentials
```

Although not recommended, credentials can also be configured directly in the helm chart. For example, add the following to your override.yaml to customize pgsql credentials:

```yaml
pgsql:
  enabled: false # disable internal pgsql database
  auth:
    database: "customdb"
    host: pgsql.database.company.com # external pgsql host
    user: "newuser"
    password: "newpassword"
    port: "5432"
```

## Using external Redis instances

The default Sourcecgraph deployment ships two separate Redis instances: [redis-cache.Deployment.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/templates/redis/redis-cache.Deployment.yaml) and [redis-store.Deployment.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/templates/redis/redis-store.Deployment.yaml).

To use external Redis instances, first review our [general recommendations](https://docs.sourcegraph.com/admin/external_services/redis). When using external Redis instances, you’ll need to specify the new endpoint for each. You can specify the endpoint directly in the values file, or by referencing an existing secret.

> An example of this approach can be found in our Helm chart [using external redis](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/external-redis) example on Github.

If your external Redis instances do not require authentication, you can configure access in your [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/examples/external-redis/override.yaml) with the `endpoint` setting:

```yaml
redisCache:
  enabled: false
  connection:
    endpoint: redis://redis-cache.example.com:6379 # use a dedicated Redis, recommended

redisStore:
  enabled: false
  connection:
    endpoint: redis://redis-shared.example.com:6379/2 # shared Redis, not recommended
```

If your endpoints do require authentication, we recommend storing the credentials in [Secrets] created outside of the helm chart and managed in a secure manner.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: redis-cache-connection
data:
  # notes: secrets data has to be base64-encoded
  endpoint: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: redis-store-connection
data:
  # notes: secrets data has to be base64-encoded
  endpoint: ""
```

You can reference this secret in your [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/examples/external-redis/override-secret.yaml) by configuring the `connection.existingSecret` key:

```yaml
redisCache:
  enabled: false
  connection:
    existingSecret: redis-cache-connection

redisStore:
  enabled: false
  connection:
    existingSecret: redis-store-connection
```

## Using external Object Storage

By default Sourcegraph uses a MinIO server bundled with the instance to temporarily store precise code intelligence indexes uploaded by users. If you prefer, you can configure your instance to store this data in an S3 or GCS bucket. To use an external Object Storage service (S3-compatible services, or GCS), first review our [general recommendations](https://docs.sourcegraph.com/admin/external_services/object_storage). Then review the following example and adjust to your use case.

To target a managed object storage service, you will need to set a handful of environment variables for configuration and authentication to the target service.

> An example of this approach can be found in our Helm chart [using external object storage](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/external-object-storage/override.yaml) example on Github.
> The example assumes the use of AWS S3. You may configure the environment variables accordingly for your own use case based on our [general recommendations](https://docs.sourcegraph.com/admin/external_services/object_storage).

If you provide credentials with an access key / secret key, we recommend storing the credentials in [Secrets] created outside of the helm chart and managed in a secure manner. An example Secret is shown here:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sourcegraph-s3-credentials
data:
  # notes: secrets data has to be base64-encoded
  PRECISE_CODE_INTEL_UPLOAD_AWS_ACCESS_KEY_ID: ""
  PRECISE_CODE_INTEL_UPLOAD_AWS_SECRET_ACCESS_KEY: ""
```

In your [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/external-object-storage/override.yaml), reference this Secret and the necessary environment variables:

```yaml

minio:
  enabled: false # Disable deployment of the built-in object storage

# we use YAML anchors and alias to keep override file clean
objectStorageEnv: &objectStorageEnv
  PRECISE_CODE_INTEL_UPLOAD_BACKEND:
    value: S3 # external object storage type, either "S3" or "GCS"
  PRECISE_CODE_INTEL_UPLOAD_BUCKET:
    value: lsif-uploads # external object storage bucket name
  PRECISE_CODE_INTEL_UPLOAD_AWS_ENDPOINT:
    value: https://s3.us-east-1.amazonaws.com
  PRECISE_CODE_INTEL_UPLOAD_AWS_REGION:
    value: us-east-1
  PRECISE_CODE_INTEL_UPLOAD_AWS_ACCESS_KEY_ID:
    secretKeyRef: # Pre-existing secret, not created by this chart
      name: sourcegraph-s3-credentials
      key: PRECISE_CODE_INTEL_UPLOAD_AWS_ACCESS_KEY_ID
  PRECISE_CODE_INTEL_UPLOAD_AWS_SECRET_ACCESS_KEY:
    secretKeyRef: # Pre-existing secret, not created by this chart
      name: sourcegraph-s3-credentials
      key: PRECISE_CODE_INTEL_UPLOAD_AWS_SECRET_ACCESS_KEY

frontend:
  env:
    <<: *objectStorageEnv

preciseCodeIntel:
  env:
    <<: *objectStorageEnv
```

## Using SSH to clone repositories

If repository authentication is required to `git clone` a repository then you must provide credentials to the container.

Create a [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) that contains the base64 encoded contents of your SSH private key and `known_hosts` file. The SSH private key should not require a passphrase. The Secret will be mounted in the `gitserver` deployment to authenticate with your code host.

**Option 1: Create Secret with Local SSH Keys**
If you have access to the ssh keys locally, you can run the command below to create the secret:

```sh
kubectl create secret generic gitserver-ssh \
	    --from-file id_rsa=${HOME}/.ssh/id_rsa \
	    --from-file known_hosts=${HOME}/.ssh/known_hosts
```

**Option 2: Create Secret from Manifest File**
Alternatively, you may manually create the secret from a manifest file.

> WARNING: For security purposes, do NOT commit the secret manifest into your Git repository unless you are comfortable storing sensitive information in plaintext and your repository is private.

Create a file with the following and save it as `gitserver-ssh.Secret.yaml`

 ***[TODO Where are they creating and storing this file? In root?]***

```sh
apiVersion: v1
kind: Secret
metadata:
  name: gitserver-ssh
data:
  # notes: secrets data has to be base64-encoded
  id_rsa: ""
  known_hosts: ""
```

Add the following values to your override file to reference the Secret:

```yaml
gitserver:
  sshSecret: gitserver-ssh
```

Apply the created Secret to your Kubernetes instance with the command below:

```sh
kubectl apply -f gitserver-ssh.Secret.yaml
```