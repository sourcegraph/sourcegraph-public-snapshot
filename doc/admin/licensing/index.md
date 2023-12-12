# Licensing

Sourcegraph requires a valid license key to enable many of its more prominent features.

License keys should not be shared across instances of Sourcegraph. If an additional license key is required for something like a dev environment, please [contact customer support](https://about.sourcegraph.com/contact) for an additional license key.

## FAQ

### What happens if our Sourcegraph license expires?

Sourcegraph will revert to a free license, and any features that require an enterprise license will stop functioning. This could lead to data loss if some of these features were in use, so be sure to renew your license in advance!

## How can we update our license key?

Any current Site Admin can update your license key by going to Site Admin -> [Site configuration](../config/site_config.md)

These settings live in the JSON object, and you will need to navigate to the _licenseKey_ section of that object.

Update the value of this with your new license key and click Save to apply your changes.

Example:
```
  "licenseKey": "<your_key_here>",
```

### We have set up a new Sourcegraph instance by replicating an existing instance, how can we generate a new site ID to ensure the instances are unique?

The site ID of a Sourcegraph instance can be updated by running the following SQL query against the database:

```sql
UPDATE global_state SET site_id = gen_random_uuid();
```

You will still require a unique license key for each site ID.
