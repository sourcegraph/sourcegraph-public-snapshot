# Sourcegraph 4.2.1+ blobstore update notes

In Sourcegraph versions 4.2.1+ and 4.3+, `minio` object storage has been replaced with `sourcegraph/blobstore`. **No migration is required.**

## If you use Kubernetes + Helm

To update a Helm deployment, simply replace `minio` with `blobstore` in your Helm override YAML file (if present). For example:

```diff
-minio:
+blobstore:
  enabled: false # Disable deployment of the built-in object storage
```

## About the change

Sourcegraph is committed to only distributing open-source software that is permitted by other industry leaders, and prohibits the use of all software licenses [not allowed at Google](https://opensource.google/documentation/reference/thirdparty/licenses#banned).

As a result, Sourcegraph v4.2.1+ and v4.3+ no longer use or distribute _any_ minio or AGPL-licensed software.

If your Sourcegraph instance is already configured to use [external object storage](../external_services/object_storage.md), or you use `DISABLE_MINIO=true` in `sourcegraph/server` deployments, then this change should not affect you (there would already be no Minio software running as part of Sourcegraph.)

If you have any questions or concerns, please reach out to us via support@sourcegraph.com and we'd be happy to help.
