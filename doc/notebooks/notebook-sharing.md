# Sharing Notebooks

Currently, Notebooks support the following sharing permissions.

## User namespace

### Private Notebooks
This is the default permissions level for all new Notebooks. Only the creator can view and edit the Notebook.

### Public Notebooks
Notebooks can be shared with everyone (public Notebooks on [sourcegraph.com](https://sourcegraph.com) are viewable by anyone and don't require a Sourcegraph account), or with your entire Sourcegraph instance

## Sourcegraph organization namespace
Find out more about Sourcegraph organizations and how to create and configure them on the [organizations docs page](../admin/organizations.md). Briefly, Sourcegraph Cloud organizations are groups of users and repositories within Sourcegraph. For on-prem or self-hosted Sourcegraph instances, organizations allow admins to specify shared settings but not repositories.

### Private organization Notebooks
Only organziation members can view and edit the Notebook.

### Public organization Notebooks
In self-hosted and managed Sourcegraph instances, everyone who has access to the instance can view the Notebook. On [sourcegraph.com](https://sourcegraph.com), anyone can view the Notebook. In both cases, only members of the owning Sourcegraph organization can edit the Notebook.
