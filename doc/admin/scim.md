# SCIM

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span> This feature is in beta and might change in the future.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

SCIM (System for Cross-domain Identity Management) is a standard for provisioning and deprovisioning users and groups in an organization. IdPs (identity providers) like Okta, OneLogin, and Azure Active Directory support provisioning users through SCIM.

Sourcegraph supports SCIM 2.0 for provisioning and de-provisioning _users_.

> WARNING: We’ve only tested the endpoint with Okta and Azure Active Directory. Be cautious when using our SCIM server with other IdPs.

## How to use

To use SCIM, you must have an existing IdP configured to connect to and authenticate with your Sourcegraph instance. For auth, we currently support Bearer token authentication.

To configure:

1. Generate a random alphanumeric bearer token of maximum 255 characters
2. Add the following line to your [site configuration](config/site_config.md):

   ```
   "scim.authToken": "{your token}"
   ```

3. Set up your IdP to use our SCIM API. The API is at

   ```
   https://sourcegraph.company.com/.api/scim/v2
   ```

   so the "Users" endpoint is at

   ```
   https://sourcegraph.company.com/.api/scim/v2/Users
   ```

## Features and limitations

### User attributes

The User endpoint only synchronizes attributes needed to create a Sourcegraph account.

We sync the following attributes:

- preferred username
- name
- email addresses

### REST methods

We support REST API calls for:

- Creating users (POST)
- Updating users (PATCH)
- Replacing users (PUT)
- Deleting users (DELETE)
- Listing users (GET)
- Getting users (GET)

### Feature support

We support the following SCIM 2.0 features:

- ✅ Updating users (PATCH)
- ✅ Pagination for listing users
- ✅ Filtering for listing users

### Limitations

- ❌ Bulk operations – need to add users one by one
- ❌ Sorting – when listing users
- ❌ Entity tags (ETags)
- ❌ Multi-tenancy – you can only have 1 SCIM client configured at a time.
- ❌ Soft delete – Currently, we do not support soft deletion through SCIM. When a user is deleted (typically, when removed from a group of users who can access Sourcegraph), we **permanently delete** their user in Sourcegraph. This means that if the user is re-added to such a group, their settings will be reset.
- ❌ Tests with many IdPs – we’ve only validated the endpoint with Okta and Azure AD.
