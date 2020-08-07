# Configuring Sourcegraph for encryption

Sourcegraph can be configured to encrypt secrets data using an encryption key created by your organization.  This allows your organization to create a secure key for encrypting all data, one that is unique to your organization. In no way can Sourcegraph, nor anyone else access your encrypted data, or encryption keys without explicit access.  

Additionally

## Creating a new encryption key

Creating an encryption key is intentional. It needs to be backed up by your organization.  Secrets are 32 characters consisting of uppercase letters, lowercase letters, and numbers.  To generate a secret key run `openssl rand -hex 32`

## Configuring Sourcegraph for encryption

There are three ways to configure Sourcegraph to encrypt your data.

1. Set the SOURCEGRAPH\_CRYPT\_KEY variable to the value of the key you created, prior to running Sourcegraph. <DAX INSERT HERE>.

1. Either create a file in */var/lib/sourcegraph/token*, or set the SOURCEGRAPH\_SECRET\_FILE variable to the location of a file that will contain the newly created encryption key. This file can be made available to the frontend docker by <DAX INSERT HERE>.  Alternatively a Kubernetes volume mount can be mounted as read only, by <DAX INSERT HERE>.

## Rotating encryption keys

Sourcegraph supports rotating encryption keys. It will transparently rotate the encryption on items, if so configured.  To enable key rotation, two different 32 character keys must be provided, separated by a comma.  The first key will be the new key used to re-encypt data, whilst the second key is the key used to decrypt data.  Keys must be stored as in the previous section.
