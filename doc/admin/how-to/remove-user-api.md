# How to remove users with the GraphQL API

This document walk you through the steps of removing users with the GraphQL API. 

## Prerequisites

* This document assumes that you have site-admin level permissions on your Sourcegraph instance
* Assumes that you have set up your query token [via the steps here](https://docs.sourcegraph.com/api/graphql#quickstart)

## Steps to remove a user

### There are two different options for removing a user:

**Option A) Deleting a user:** the user and *all* associated data is marked as deleted in the DB and never served again. You could undo this by running DB commands manually.

**Option B) Nuking a user:** the user and *all* associated data is deleted forever. *Note: You cannot undo this and this is considered the less safe option.*

First, query the user's ID from their email address or user name

Example:

```json
{
  user(email: "someone@gmail.com") {
    id
  }
}
```

Next, plug the user ID into either Option A. or B. to delete the account.

Option A) example:

```json
mutation {
  deleteUser(user: "THE_USER_ID") {
    alwaysNil
  }
}
```

Option B) example, include `hard: true`:

```json
mutation {
  deleteUser(user: "THE_USER_ID" hard: true) {
    alwaysNil
  }
}
```

**Optional step:** Recheck the delete worked by running the query from step #1. again. If the results are “user not found:...” then it worked. 

Example:

```json
{
  "errors": [
    {
      "message": "user not found: [someone@gmail.com]",
      "path": [
        "user"
      ]
    }
  ],
  "data": {
    "user": null
  }
}
```


## Further resources

* [Sourcegraph - User data deletion](https://docs.sourcegraph.com/admin/user_data_deletion)
* [Sourcegraph GraphQL API](https://docs.sourcegraph.com/api/graphql)
