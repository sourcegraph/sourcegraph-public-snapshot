# Pings

Sourcegraph periodically sends a ping to Sourcegraph.com to help our product and customer teams. It sends only the high-level data below. It never sends code, repository names, usernames, or any other specific data. To learn more, go to the **Site admin > Pings** page on your instance. (The URL is `https://sourcegraph.example.com/site-admin/pings`.)

- Sourcegraph version string
- Deployment type (single-node or Kubernetes cluster)
- Whether the instance is deployed on localhost (true/false)
- Randomly generated site identifier
- The email address of the initial site installer (or if deleted, the first active site admin), to know who to contact regarding sales, product updates, and policy updates
- Which category of authentication provider is in use (built-in, OpenID Connect, an HTTP proxy, SAML, GitHub, GitLab)
- Which code hosts are in use (GitHub, Bitbucket Server, GitLab, Phabricator, Gitolite, AWS CodeCommit, Other)
- Whether new user signup is allowed (true/false)
- Whether a repository has ever been added (true/false)
- Whether a code search has ever been executed (true/false)
- Whether code intelligence has ever been used (true/false)
- Total count of existing user accounts
- Aggregate counts of current daily, weekly, and monthly users
- Aggregate counts of current users using code host integrations
- Aggregate counts of current users by product feature (site management, code search and navigation, code review, saved searches, diff searches)
- Aggregate daily, weekly, and monthly latencies (in ms) of code intelligence events (e.g., hover tooltips) and search queries
- Aggregate daily, weekly, and monthly total counts of code intelligence events (e.g., hover tooltips)
- Total count of code campaigns created
- Aggregate counts of current daily, weekly, and monthly users using search
- Aggregate daily, weekly, and monthly total users using each search mode
- Aggregate daily, weekly, and monthly total searches using each search mode
- Aggregate daily, weekly, and monthly total counts of searches using each search filter.
- Aggregate daily, weekly, and monthly total counts of users conducting searches using each search filter.

To disable pings, please [contact support](https://about.sourcegraph.com/contact/).
