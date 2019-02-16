# Bitbucket Server

Site admins can sync Git repositories hosted on [Bitbucket Server](https://www.atlassian.com/software/bitbucket/server) (and the [Bitbucket Data Center](https://www.atlassian.com/enterprise/data-center/bitbucket) deployment option) with Sourcegraph so that users can search and navigate the repositories.

To set this up, add Bitbucket Server as an external service to Sourcegraph:

1. Go to **User menu > Site admin**.
1. Open the **External services** page.
1. Press **+ Add external service**.
1. Enter a **Display name** (using "Bitbucket Server" is OK if you only have one Bitbucket Server instance).
1. In the **Kind** menu, select **Bitbucket Server**.
1. Configure the connection to Bitbucket Server in the JSON editor. Use Cmd/Ctrl+Space for completion, and [see configuration documentation below](#configuration).
1. Press **Add external service**.

## Repository syncing

### Authentication for older Bitbucket Server versions

Bitbucket Server versions older than v5.5 require specifying a less secure username and password combination, as those versions of Bitbucket Server do not support [personal access tokens](https://confluence.atlassian.com/bitbucketserver/personal-access-tokens-939515499.html).

### Excluding personal repositories

Sourcegraph will be able to view and clone the repositories that the account you provide has access to. For example, if you provide a personal access token or username/password of an administrator Bitbucket Server account, Sourcegraph will be able to view and clone all repositories -- even personal ones.

We recommend that you create a new Bitbucket user account specifically for Sourcegraph (e.g. a "Sourcegraph Bot" account) and only give that account access to the repositories you wish to be viewable on Sourcegraph.

If you don't wish to create a separate Bitbucket user account just for Sourcegraph, you can specify `"excludePersonalRepositories": true` in the configuration. With this enabled, Sourcegraph will exclude any personal repositories from being imported, even if it has access to them.

### HTTPS cloning

Sourcegraph by default clones repositories from your Bitbucket Server via HTTP(S), using the access token or account credentials you provide in the configuration. SSH cloning is not used, so you don't need to configure SSH cloning.

## Configuration

Bitbucket Server external service connections support the following configuration options, which are specified in the JSON editor in the site admin external services area.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/bitbucket_server.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/bitbucket_server) to see rendered content.</div>
