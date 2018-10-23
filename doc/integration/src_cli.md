# `src` CLI

The [`src` CLI tool](https://github.com/sourcegraph/src-cli) provides access to Sourcegraph via a command-line interface.

# Browser authorization flow for clients

It's now easier to use the Sourcegraph [`src` CLI](https://github.com/sourcegraph/src-cli) and other client applications that require authorization with your Sourcegraph instance. Sourcegraph now supports a browser-based authorization flow, so you can just approve the request in your browser (and don't need to manually generate an access token).

Here's how it looks for the `src` CLI:

1. Run `src init` to open your web browser to the Sourcegraph browser-based authorization flow for Sourcegraph.com.

   Run `src init --url https://sourcegraph.example.com` for a self-hosted Sourcegraph instance.
   
   If you're in a terminal with no web browser, use the `--console-only` flag to print the URL to open manually.
1. Follow the prompts to sign in (if needed).
1. Review and approve the authorization request.

When finished, the `src` CLI will be authorized (with an access token that you can revoke at any time in your user account area). Try running `src config get` to get your user settings, for example.

If you've used the [`gcloud init` command](https://cloud.google.com/sdk/docs/initializing) with Google Cloud, this process will be familiar.
