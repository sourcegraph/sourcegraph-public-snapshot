# Licensing

Sourcegraph Enterprise requires a valid license key to enable many of the Enterprise-specific features.

In order to ensure that a valid license is configured, Sourcegraph periodically does a verification check to Sourcegraph.com. This check sends no information other than what's necessary to verify your Sourcegraph license. If this check fails, or if the check does not occur, Sourcegraph will disable all Enterprise features until a successful license check is completed.

If this behaviour is undesired, or Sourcegraph has to operate in an environment without an external internet connection, contact customer support for a special license key.

## Troubleshooting

### A customer has the same site ID on multiple instances

A Sourcegraph site ID should be unique. However, if a customer were to exactly clone a Sourcegraph instance, it could happen that the site ID is shared among more than one instance. In this case, a new site ID must be created for one of the instances.

A customer can update their site ID by running the following query:

```sql
UPDATE global_state SET site_id = gen_random_uuid();
```

A customer will still require a unique license key for each site ID.
