# Deployment Repositories

## Create a private copy

### Step 1: Create an empty repository

Follow the [official GitHub docs](https://docs.github.com/en/repositories/creating-and-managing-repositories/creating-a-new-repository) on creating a new **empty** repository.

### Step 2: Set the environment variables

Export the following environment variables for the next steps.

- `DEPLOY_REPO_NAME`: name of the deployment repository
  - `deploy-sourcegraph` for Kubernetes with Kustomize deployment
  - `deploy-sourcegraph` for Docker and Docker Compose deployment
- `DEPLOY_GITHUB_USERNAME`: the account name that is hosting the empty repository created in step 1 
- `PRIVATE_DEPLOY_REPO_NAME`: default to the same name as $DEPLOY_REPO_NAME
- `SOURCEGRAPH_VERSION`: latest version number of Sourcegraph

Update the environment variables in the command below before running it in your terminal:

```bash
export DEPLOY_GITHUB_USERNAME="YOUR_USERNAME"
export DEPLOY_REPO_NAME="deploy-sourcegraph"
export PRIVATE_DEPLOY_REPO_NAME="$DEPLOY_REPO_NAME"
export SOURCEGRAPH_VERSION="v4.3.0"
```

### Step 3: Create remote and local copies

Once the required environment variables are exported, run the following commands in the same terminal:

```bash
git clone --bare https://github.com/sourcegraph/$DEPLOY_REPO
cd $DEPLOY_REPO.git
git push --mirror https://github.com/$DEPLOY_GITHUB_USERNAME/$PRIVATE_DEPLOY_REPO_NAME.git
cd ..
rm -rf $DEPLOY_REPO.git
git clone https://github.com/$DEPLOY_GITHUB_USERNAME/$PRIVATE_DEPLOY_REPO_NAME.git
cd $PRIVATE_DEPLOY_REPO_NAME

```

### Step 4: Create a release branch

Create a `release` branch to track all of your customizations to Sourcegraph. This branch will be used to [upgrade Sourcegraph](update.md) and [install your Sourcegraph instance](./index.md#installation).

```bash
git checkout $SOURCEGRAPH_VERSION -b release-$SOURCEGRAPH_VERSION
```
