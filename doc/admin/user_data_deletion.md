# User data deletion

As a site administrator, you have the ability to delete users and their associated data on the **Admin** -> **Users** page (https://sourcegraph.example.com/site-admin/users).

On this page, you are presented two options:

- "Delete": the user and ALL associated data is marked as deleted in the DB and never served again.
- "Delete forever": the user and ALL associated data is permanently removed from the DB (you CANNOT undo this).

When deleting a user with either option, the following information is removed:

- All user data (access tokens, email addresses, external account info, survey responses, etc)
- Organization membership information (which organizations the user is a part of, any invitations created by or targeting the user).
- Sourcegraph extensions published by the user on the instance the deletion request is sent to.
- User, Organization, or Global settings authored or modified by the user.
- Bulk operations on changesets and batch changes created by the user.

> WARNING: If deleted and recreated a user will no longer have access to the configurations and customizations they may have made for their old account. The database will store information about their previous configuration, but a user associated with a new `id` in the `users` table in PostgreSQL will be created. For example, site admin privileges will have to be re-assigned if a once site admin user is deleted and recreated. 
