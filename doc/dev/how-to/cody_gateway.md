# How to set up Cody Gateway locally

> WARNING: This is a development guide - to use Cody Gateway for Sourcegraph, refer to [Sourcegraph Cody Gateway](../../cody/core-concepts/cody-gateway.md).

This guide documents how to set up [Cody Gateway](https://handbook.sourcegraph.com/departments/engineering/teams/cody/cody-gateway/) locally for development.

To get started, Cody Gateway is included in the standard `dotcom` run set.
Since Cody Gateway cucrently depends on Sourcegraph.com, there's not much point running any other run set.

```sh
sg start dotcom
```

Then, set up some feature flags:

- `product-subscriptions-service-account`: set to `true` globally for convenience.
  In production, this flag is used to denote service accounts, but in development it doesn't matter.
  - You can also create an additional user and set this flag to `true` only for that user for more robust testing.

Configure Cody features to talk to your local Cody Gateway:

```json
{
  "completions": {
    "enabled": true,
    "provider": "sourcegraph",
    "endpoint": "http://localhost:9992",
    "chatModel": "anthropic/claude-2",
    "completionModel": "anthropic/claude-instant-1",
    // Create a subscription and create a license key:
    // https://sourcegraph.test:3443/site-admin/dotcom/product/subscriptions
    // Under "Cody services", ensure access is enabled and get the access token
    // to use here.
    // Note that the license and tokens will only work locally.
    "accessToken": "..."
  }
}
```

Add the following to your `sg.config.overwrite.yaml`:

```yaml
commands:
  cody-gateway:
    env:
      # Get from dev-private access token
      CODY_GATEWAY_ANTHROPIC_ACCESS_TOKEN: "..."
      # Create a personal access token on https://sourcegraph.test:3443/user/settings/tokens
      # or on your `product-subscriptions-service-account` user.
      CODY_GATEWAY_DOTCOM_ACCESS_TOKEN: "..."
```

For more configuration options, refer to the [configuration source code](https://github.com/sourcegraph/sourcegraph/blob/main/cmd/cody-gateway/shared/config.go#L60).

Then, restart `sg start dotcom` and try interacting with Cody!

## Observing Cody Gateway

You can get full tracing of Cody interactions, including Cody Gateway parts, with `sg start monitoring`'s Jaeger component.
Cody Gateway will also emit traces of its background jobs.

All event logging is output as standard logs in development under the `cody-gateway.events` log scope.

## Working with BigQuery for events logging

> NOTE: To make authentication work magically with the GCP, you need to install the [`gcloud`](https://cloud.google.com/sdk/docs/install-sdk) CLI and
run `gcloud auth login --project cody-gateway-dev` once.

To send events to BigQuery while developing Cody Gateway locally, add the following to your `sg.config.overwrite.yaml`:

```yaml
commands:
  cody-gateway:
    env:
      CODY_GATEWAY_BIGQUERY_PROJECT_ID: cody-gateway-dev
```

Then to view events statistics on the product subscription page, add the following section in the site configuration, and run the `sg start dotcom` stack:

```json
{
  "dotcom": {
    "codyGateway": {
      "bigQueryGoogleProjectID": "cody-gateway-dev",
      "bigQueryDataset": "cody_gateway",
      "bigQueryTable": "events"
    }
  }
}
```
