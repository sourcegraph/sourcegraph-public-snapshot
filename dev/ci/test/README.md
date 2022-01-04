# Running e2e/qa tests locally

## Requirements

Before running tests export these environment variables:

```
export LOG_STATUS_MESSAGES=true
export NO_CLEANUP=false
export SOURCEGRAPH_SUDO_USER=admin
export SOURCEGRAPH_BASE_URL="http://127.0.0.1:7080"
export TEST_USER_EMAIL="test@sourcegraph.com"
export TEST_USER_PASSWORD="supersecurepassword"
export INCLUDE_ADMIN_ONBOARDING="false"
# Set the IMAGE to whichever version of Sourcegraph you want to test
export IMAGE="sourcegraph/server:insiders"
# Set the following to a valid github token. Your personal github token should have access to all the repos in the Sourcegraph github required to run these tests.
export GITHUB_TOKEN=<insert token here>
```

## Running tests

### E2E

From the root of this repository:

```
./dev/ci/e2e.sh
```

### QA

From the root of this repository

1.

```
CLEAN="true" ./dev/run-server-image.sh -d --name sourcegraph
```

1. Login to the instance at `http://locahost:7080` and create a user with the following details.

```
email=test@sourcegraph.com
user=admin
password=supersecurepassword
```

1. Create an access token with admin access, copy the value and export it as follows:

```
export SOURCEGRAPH_SUDO_TOKEN=<insert token here>
```

1. Run the QA tests as follows.

```
cd client/web
yarn run test:regression
```

### Codeintel QA

TODO: add instrucions to run codeintel QA locally
