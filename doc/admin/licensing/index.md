# Licensing

[Sourcegraph Enterprise](../../getting-started/oss-enterprise.md) requires a valid license key to enable many Enterprise-specific features.

Sourcegraph will periodically perform a license validation check by contacting sourcegraph.com. This check sends no information other than a unique site ID and information about the configured Sourcegraph license. This check is mandatory, and if the check fails, or if the check does not occur, Sourcegraph will disable all Enterprise features until a successful license check is completed.

If Sourcegraph needs to operate in a sandboxed environment without an external internet connection, contact customer support for a special license key.

License keys also need to be unique to a single instance of Sourcegraph. If the same license key is used across multiple instances, subsequent license checks will fail. If multiple license keys are required for dev/staging instances, contact customer support for additional license keys for these instances.

## Troubleshooting

### We have set up a new Sourcegraph instance by replicating an existing instance, how can we generate a new site ID to ensure the instances are unique?

The site ID of a Sourcegraph instance can be updated by running the following SQL query against the database:

```sql
UPDATE global_state SET site_id = gen_random_uuid();
```

You will still require a unique license key for each site ID.
