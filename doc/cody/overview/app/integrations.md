## Cody app API integrations

This page outlines how extensions and other clients can integrate with Cody app.

## API

The Cody app API is a GraphQL API exposed on the localhost port which is `http://localhost:3080` by default.

## App.json file

The `app.json` file is placed in the app's local data directory when it runs for the first time and it provides the token and local endpoint URL that local clients can use to connect with the app's API.

The location of the file varies by OS:

- Mac OS: `~/Library/Application Support/com.sourcegraph.cody/app.json`
- Linux:
 - `$XDG_DATA_HOME/com.sourcegraph.cody/app.json`
 - or: `$HOME/.local/share/com.sourcegraph.cody/app.json`
- Windows: `{FOLDERID_LocalAppData}/com.sourcegraph.cody/app.json`

Example contents of `app.json`:

```json
{
  "token": "xxxxx",
  "endpoint": "http://localhost:3080",
  "version": "2023.06.16"
}
```

## Token

API requests require a token. The token is generated when the Cody app runs for the first time and placed in `app.json` in the app's config directory. See the [Sourcegraph API docs](../../../api/graphql/index.md) for how to make requests with the token.

## Endpoint

The local API endpoint is `http://localhost:3080` by default. The `endpoint` field in the `app.json` file provides the endpoint that the current app installation is using.

## Detecting if the Cody app is installed

To detect if the Cody app has been installed locally and run at least once, check for the existance of the `app.json` file in the expected directory

## Detecting if the Cody app is running

To detect if the Cody app is running a GET request to the version endpoint can be used.  The endpoint can be found at `/__version` and if the app is running it will return a string that is the current version.
## Deep links

The Cody app supports opening deep links that start with the `sourcegraph://` prefix. Opening of a deep links from a browser or from another application will launch the app if it's not already running and will navigate the app's UI to the destination.
