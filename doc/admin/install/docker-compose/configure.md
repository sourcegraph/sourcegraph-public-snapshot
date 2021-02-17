# Configuring Sourcegraph

## Fork this repository

We **strongly** recommend that you create your own fork of [sourcegraph/deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/) to track customizations to the [Sourcegraph Docker Compose yaml](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml). This will make upgrades far easier.

- Fork [sourcegraph/deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/)
  - The fork can be public **unless** you plan to store secrets (SSL certificates, external Postgres credentials, etc.) in the repository itself.

- Create a `release` branch to track all of your customizations to Sourcegraph. When you upgrade Sourcegraph, you will merge upstream into this branch.

```bash
SOURCEGRAPH_VERSION="v3.24.1"
git checkout $SOURCEGRAPH_VERSION -b release
```

- Commit customizations to the [Sourcegraph Docker Compose yaml](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) to your `release` branch

- When you upgrade, merge the corresponding upstream release tag into your release branch. E.g., `git remote add upstream https://github.com/sourcegraph/deploy-sourcegraph` to add the upstream remote and `git checkout release && git merge v3.15.0` to merge the upstream release tag into your release branch.
