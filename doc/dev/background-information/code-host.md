# Code host connections on local dev environment

Connecting to a code host is required to develop and test certain features, including:

- Adding public and private repositories
- Creating and modifying search contexts

To enable code host connections, there are two options: enabling site-wide or
enabling for a specific user.

## Enabling site-wide (enabled by default on local dev environment)

Code host connections can be enabled site-wide by adding the following to the
site config JSON file:

```json
"externalService.userMode": "all"
```

## Enabling for specific users

To enable code host connections for a user, two flags need to be set:
`AllowUserExternalServicePublic` and `AllowUserExternalServicePrivate`.
These flags may not be set when testing locally. You can run the following
SQL command to set the flags for all users in your local test instance:

```sql
update users 
set tags = '{AllowUserExternalServicePublic,AllowUserExternalServicePrivate}'
where id={USER ID HERE};
```
