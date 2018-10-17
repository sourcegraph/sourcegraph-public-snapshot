# Pings

Sourcegraph periodically sends a ping to Sourcegraph.com to help our product and customer teams. It sends only the high-level data below. It never sends code, repository names, usernames, or any other specific data. To learn more, click **Admin** in the top right of any page on your instance, and go to **Pings** in the left side menu. (The URL is `https://sourcegraph.example.com/site-admin/pings`.)

- Sourcegraph version string
- Deployment type (single-node or Kubernetes cluster)
- Randomly generated site identifier
- The email address of the initial site installer (or if deleted, the first active site admin), to know who to contact regarding sales, product updates, and policy updates
- Which category of authentication provider is in use (built-in, OpenID Connect, an HTTP proxy, or SAML)
- Whether code intelligence is enabled
- Total count of existing user accounts
- Aggregate counts of current daily, weekly, and monthly users
- Aggregate counts of current users using code host integrations
