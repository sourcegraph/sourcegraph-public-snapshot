# User data deletion

As a site administrator, you have the ability to delete users and their associated data on the **Admin** -> **Users** page (https://sourcegraph.example.com/site-admin/users).

On this page, you are presented two options:

- Deleting a user: the user and ALL associated data is marked as deleted in the DB and never served again. You could undo this by running DB commands manually.
- Nuking a user, the user and ALL associated data is deleted forever (you CANNOT undo this). For GDPR-style requests, nuking is used (but not sufficient on its own, see below).

When deleting or nuking a user, the following information is removed:

- All user data (access tokens, email addresses, external account info, survey responses, etc)
- Organization membership information (which organizations the user is a part of, any invitations created by or targeting the user).
- Sourcegraph Extensions published by the user.
- User, Organization, or Global settings authored or modified by the user.
- Discussion threads and comments created by the user.

Data that is NOT currently removed:

- BUG(@dadlerj): Redis store user activity data
- BUG(@dadlerj): CRM, Analytics DB, etc.
- Repositories the user may have added (impossible because we don't track this).
- Organizations the user may have created (impossible because we don't track this, and it would evict other members).
- Extension releases the user may have made on extensions _not created by that same user_.
- Product licenses & subscriptions the user has created (this would prevent us from having a legal record of sales).

For GDPR-style deletion requests, the above user data not deleted as part of this operation and must currently be performed manually. Contact support@sourcegraph.com for more information.
