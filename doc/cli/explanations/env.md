# Environment variables

## Overview

`src` requires two environment variables to be set to authenticate against your Sourcegraph instance.

## `SRC_ENDPOINT`

`SRC_ENDPOINT` defines the base URL for your Sourcegraph instance. In most cases, this will be a simple HTTPS URL, such as the following:

```
https://sourcegraph.com
```

If you're unsure what the URL for your Sourcegraph instance is, please contact your site administrator.

## `SRC_ACCESS_TOKEN`

`src` uses an access token to authenticate as you to your Sourcegraph instance. This token needs to be in the `SRC_ACCESS_TOKEN` environment variable.

To create an access token, please refer to "[Creating an access token](../how-tos/creating_an_access_token.md)".
