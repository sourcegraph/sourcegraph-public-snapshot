# Sourcegraph Instance Validation

>🚨 WARNING 🚨: **Sourcegraph Instance Validation is currently experimental.** We're exploring this feature set. 
>Let us know what you think! [File an issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose)
>with feedback/problems/questions, or [contact us directly](https://about.sourcegraph.com/contact).

Instance validation provides a quick way to check that a Sourcegraph instance functions properly after a fresh install
 or an update.

The [`src` CLI](https://github.com/sourcegraph/src-cli) has an experimental command `validate install` which drives the
 validation from a user-provided configuration file with a validation specification (in JSON or YAML format). If no validation specification file is provided it will execute the following: 
 
* temporarily adds an external service
* waits for a repository to be cloned
* performs a search on the cloned repo
* performs a search on a non-indexed branch of the cloned repo
* creates basic code insight
* removes the added external service
* removes the added code insight

### Validation specification
 
Validation specifications can be provided in either a YAML or JSON format. The best way to describe this initial, simple, and experimental validation specification is with the example below:

#### YAML File Specification

```yaml
# adds the specified code host
externalService:
  kind: GITHUB
  displayName: srcgraph-test
  # set to true if this code host config should be deleted at the end of validation
  deleteWhenDone: true
  # maxRetries amount of retries for cloning repo
  maxRetries: 5
  # retryTimeoutSeconds wait in seconds between retries
  retryTimeoutSeconds: 5
  config:
    gitHub:
      url: https://github.com
      orgs: []
      repos:
        - sourcegraph-testing/zap

# performs the specified search and checks that at least one result is returned
searchQuery: 
  - repo:^github.com/sourcegraph-testing/zap$ test
  - repo:^github.com/sourcegraph-testing/zap$@v1.14.1 test

# checks to see if instance can create code insights
insight:
  title: "Javascript to Typescript migration"
  dataSeries:
    [ {
      "query": "lang:javascript",
      "label": "javascript",
      "repositoryScope": [
        "github.com/sourcegraph/sourcegraph"
      ],
      "lineColor": "#6495ED",
      "timeScopeUnit": "MONTH",
      "timeScopeValue": 1
    },
      {
        "query": "lang:typescript",
        "label": "typescript",
        "lineColor": "#DE3163",
        "repositoryScope": [
          "github.com/sourcegraph/sourcegraph"
        ],
        "timeScopeUnit": "MONTH",
        "timeScopeValue": 1
      }
    ]
```
#### JSON File Specification

```json
{
  "externalService": {
    "kind": "GITHUB",
    "displayName": "srcgraph-test",
    "deleteWhenDone": true,
    "maxRetries": 5,
    "retryTimeoutSeconds": 5,
    "config": {
      "gitHub": {
        "url": "https://github.com",
        "orgs": [],
        "repos": [
          "sourcegraph-testing/zap"
        ]
      }
    }
  },

  "searchQuery": [
    "repo:^github.com/sourcegraph-testing/zap$ test",
    "repo:^github.com/sourcegraph-testing/zap$@v1.14.1 test"
  ],

  "insight": {
    "title": "Javascript to Typescript migration",
    "dataSeries": [
      {
        "query": "lang:javascript",
        "label": "javascript",
        "repositoryScope": [
          "github.com/sourcegraph/sourcegraph"
        ],
        "lineColor": "#6495ED",
        "timeScopeUnit": "MONTH",
        "timeScopeValue": 1
      },
      {
        "query": "lang:typescript",
        "label": "typescript",
        "lineColor": "#DE3163",
        "repositoryScope": [
          "github.com/sourcegraph/sourcegraph"
        ],
        "timeScopeUnit": "MONTH",
        "timeScopeValue": 1
      }
    ]
  }

}
```

With this configuration, the validation command executes the following steps: 

* adds an external service
* waits for a repository to be cloned
* performs a search
* removes the external service
* creates a code insight
* removes the code insight
 
>NOTE: Every step is optional (if the corresponding top-level key is not present then the step is skipped).
> 

### Passing in secrets

Secrets are handled via environment variables. `src validate install` requires an environment variables to be set to authenticate to your code host.

>NOTE: Currently only GitHub is supported as a code host.
>

`src validate install` requires an environment variable to be setup to authenticate against your code host.

* `SRC_GITHUB_TOKEN` - defines the GitHub access token used to authenticate to the GitHub API. This must be set to use a GitHub code host.


Example:
```bash
export SRC_GITHUB_TOKEN=token
```

### Usage

Use the [`src` CLI](https://github.com/sourcegraph/src-cli) to validate with a validation specification file:
```shell script
src validate install validate.yaml
```
```shell script
src validate install validate.json
```
To execute default validation checks:

```shell script
src validate install
```

The `src` binary finds the Sourcegraph instance to validate from the environment variables 
[`SRC_ENDPOINT` and `SRC_ACCESS_TOKEN`](https://github.com/sourcegraph/src-cli#setup-with-your-sourcegraph-instance). 
