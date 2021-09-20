# Code host connections on local dev environment

Connecting to a code host is required to develop and test certain features, including:

- Adding public and private repositories
- Creating and modifying search contexts

To enable code host connections, two flags need to be set for a user:
`AllowUserExternalServicePublic` and `AllowUserExternalServicePrivate`.
These flags may not be set when testing locally. You can run the following
SQL command to set the flags for all users in your local test instance:

```sql
update users set tags = '{AllowUserExternalServicePublic,AllowUserExternalServicePrivate}';
```
