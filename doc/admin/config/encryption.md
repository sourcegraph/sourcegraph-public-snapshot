# On-disk database encryption

Sourcegraph supports encryption of sensitive data via the `encryption.keys` config option. It supports multiple encryption backends for different use cases & environments.

Currently supported encryption backends:
* Google Cloud KMS
* Mounted key (env var or file) AES encryption

## Enabling
To enable encryption you must specify key config for each of the keys defined in `encryption.keys`. You can specify the same key for all keys if you choose to, but you must at least specify config for all of them.

```json
{
  "encryption.keys": {
    // encrypts data in external_services
    "externalServiceKey": {
      "type": "mounted", // use the mounted AES encryption key
      "filePath": "/path/to/my/encryption.key" // path to a file containing your secret key
    },
    // encrypts data in user_external_accounts
    "userExternalAccountKey": {
      "type": "cloudkms", // use Google Cloud KMS
      "keyname": "/projects/my-project/name/of/my/keyring/cryptoKeys/key", // the resource name of your encryption key
      "credentialsFile": "/path/to/my/service-account.json" // path to a service account key file with the encrypter/decrypter & key viewer roles
    }
  }
}
```


## Migration
When you first enable encryption two migrations will begin in the UI (https://sourcegraph.example.com/site-admin/migrations) called 'Encrypt auth data' and 'Encrypt configuration'. These jobs watch the site config waiting for a key to be configured and then iterate over all data in the relevant tables & encrypt it. Once these two migrations reach 100% your data will be fully encrypted! You can still use sourcegraph whilst these migrations are progressing, any unencrypted data will be read as normal, and encrypted if you update it.

## Key rotation
If you use the Google Cloud KMS backend (or other future API based encryption backend) key rotation will be handled for you by the API. Currently key rotation is not supported in the 'mounted key' backend.

## Disabling encryption
If you decide to disable encryption, or want to switch to a new key, you must first decrypt the database. In order to do this you have to do a few things:
* set the env var `ALLOW_DECRYPT_MIGRATION` to `true` on the `frontend` service. this allows decryptions to run, and prevents arbitrary queries decrypting the database.
* update the migration to run the decryption (see SQL below).
* once the migrations are complete, remove the keys from site config.

Bear in mind that any new writes during these migrations will still be encrypted, the migrations will automatically pick them up & decrypt them.

In order to reverse the migration you should update the migrations to set `apply_reverse` to true:
```sql
    -- for the external_services migration
    UPDATE out_of_band_migrations SET apply_reverse = true WHERE id = 3;

    -- for the user_external_accounts migration
    UPDATE out_of_band_migrations SET apply_reverse = true WHERE id = 6;
``` 
