# GraphQL integration tests

This directory contains API-based integration tests in the form of standard Go tests. It is called gqltest since most of our API is GraphQL. However, the test suite has been extended to test other endpoints such as streaming search.

## How to set up credentials

Tests use environment variables to accept credentials of different external services involved, it is suggested to use [direnv](https://direnv.net/) to persist those credentials for your convenience. Here is a comprehensive example `.envrc` file (you're free to use any other means, e.g. `.profile` or `.bashrc`):

```sh
# Your GitHub personal access token, this token needs to have scope to access private
# repositories of "sgtest" organization on `ghe.sgdev.org`. If you haven't joined "sgtest" organization,
# please post a message on #dev-chat to ask for an invite.
export GITHUB_TOKEN=<REDACTED>

# Please go to https://team-sourcegraph.1password.com/vaults/dnrhbauihkhjs5ag6vszsme45a/allitems/zpxz7vl3ek7j3yxbnjvh6utrei
# and copy relevant credentials to here.
export AWS_ACCESS_KEY_ID=<REDACTED>
export AWS_SECRET_ACCESS_KEY=<REDACTED>
export AWS_CODE_COMMIT_USERNAME=<REDACTED>
export AWS_CODE_COMMIT_PASSWORD=<REDACTED>

export BITBUCKET_SERVER_URL=<REDACTED>
export BITBUCKET_SERVER_TOKEN=<REDACTED>
export BITBUCKET_SERVER_USERNAME=<REDACTED>

export PERFORCE_PORT=<REDACTED>
export PERFORCE_USER=<REDACTED>
export PERFORCE_PASSWORD=<REDACTED>
```

You need to run `direnv allow` after editing the `.envrc` file (it is suggested to place the `.envrc` file under `dev/gqltest`).

Alternatively you can use the 1password CLI tool:

```sh
# dev-private token for ghe.sgdev.org
op get item bw4nttlfqve3rc6xqzbqq7l7pm | jq -r '.. | select(.t? == "k8s.sgdev.org") | @sh "export GITHUB_TOKEN=\(.v)"'
# AWS and Bitbucket tokens
op get item 5q5lnpirajegt7uifngeabrak4 | jq -r '.details.sections[] | .fields[] | @sh "export \(.t)=\(.v)"'
```

## How to run tests

GraphQL-based integration tests are running against a live Sourcegraph instance, the easiest way to make one is by booting up a single Docker container:

```sh
# For easier testing, run Sourcegraph instance without volume,
# so it always starts from a clean state.
docker run --publish 7080:7080 --rm sourcegraph/server:insiders
```

Once the instance is live (look for the log line `✱ Sourcegraph is ready at: http://127.0.0.1:7080`), you can open another terminal tab to run these tests under this directory (`dev/gqltest`):

```sh
→ go test -long
2020/07/17 14:17:32 Site admin has been created: gqltest-admin
PASS
ok  	github.com/sourcegraph/sourcegraph/dev/gqltest	31.521s
```

### Testing against local dev instance

It is not required to boot up a single Docker container to run these tests, which means it's also possible to run these tests against any Sourcegraph instance, for example, your local dev instance:

```sh
go test -long -base-url "http://localhost:3080" -email "joe@sourcegraph.com" -username "joe" -password "<REDACTED>"
```

You will need to run your local instance in `enterprise` mode in order for tests to pass. Also note you should not use an external service config file. To ensure your local environment is set up correctly, follow these steps:

1. Clear your database: `./dev/drop-entire-local-database-and-redis.sh`
2. Delete your `~/.sourcegraph` directory
3. Add the following to your `sg.config.overwrite.yaml`

```yaml
commands:
  enterprise-frontend:
    env:
      EXTSVC_CONFIG_FILE: ''
    watch:
      - lib
      - internal
      - cmd/frontend
      - enterprise/internal
      - enterprise/cmd/frontend
```

4. Start your instance by running `sg start enterprise`
5. Create the admin account so that it matches the credentials passed to tests as above. (If you cleared your database this is done automatically when tests are first run)

Generally, you're able to repeatedly run these tests regardless of any failures because tests are written in the way that cleans up and restores to the previous state. It is aware of if the instance has been initialized, so you can focus on debugging tests.

Because we're using the standard Go test framework, you are able to just run a single or subset of these tests:

```sh
→ go test -long -run TestSearch
2020/07/17 14:20:59 Site admin authenticated: gqltest-admin
PASS
ok  	github.com/sourcegraph/sourcegraph/dev/gqltest	3.073s
```

## How to add new tests

Adding new tests to this test suite is as easy as adding a Go test, here are some general rules to follow:

- Use `gqltest-` prefix for entity name, and be as specific as possible for easier debugging, e.g. `gqltest-org-user-1`.
- Restore the previous state regardless of failures, including:
  - Delete new users created during the test.
  - Delete external services created during the test.
  - Although, sometimes you would not want to delete an entity so you could login and inspect the failure state.
- Prefix your branch name with `backend-integration/` will run integration tests in CI on your pull request.
