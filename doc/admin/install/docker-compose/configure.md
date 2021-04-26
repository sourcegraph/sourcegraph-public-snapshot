# Configuring Sourcegraph

Configuring Sourcegraph deployed with Docker-Compose.

## Fork this repository

We **strongly** recommend that you create your own fork of [sourcegraph/deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/) to track customizations to the [Sourcegraph Docker Compose yaml](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml). **This will make upgrades far easier.**

- Fork the [sourcegraph/deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/) repository
  > WARNING: Set it to **private** if you plan to store secrets (SSL certificates, external Postgres credentials, etc.) within the repository.

- Create a `release` branch to track all of your customizations to Sourcegraph.
  > NOTE: When you upgrade Sourcegraph, you will merge upstream into this branch.

```bash
export SOURCEGRAPH_VERSION="v3.26.3"
git checkout $SOURCEGRAPH_VERSION -b release
```

- Commit customizations to the [Sourcegraph Docker Compose yaml](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) to your `release` branch

## Upgrading with a forked repository

- When you upgrade, merge the corresponding `upstream release` tag into your `release` branch. 

```bash
# to add the upstream remote.
git remote add upstream https://github.com/sourcegraph/deploy-sourcegraph-docker
# to merge the upstream release tag into your release branch.
git checkout release && git merge v3.26.3
```