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

To enable code host connections for a user, one of two flags need to be set:
`AllowUserExternalServicePublic` and `AllowUserExternalServicePrivate`.
The former allows users to add public repositories, while the latter allows
users to add both public and private repositories.
These flags may not already be set when testing locally. You can run the following
SQL command to set the flags for a user in your local test instance
(replace `{USER ID HERE}` with the user ID):

```sql
update users
set tags = array_append(tags, 'AllowUserExternalServicePrivate')
where not ('AllowUserExternalServicePrivate'=any(tags)) and id = {USER ID HERE};
```
