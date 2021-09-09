# Web Application

## Local development

Use `sg` CLI tool to configure and start local development server. For more information checkout `sg` [README](../../dev/sg/README.md).

### Configuration

Environment variables important for the web server:

1. `WEBPACK_SERVE_INDEX` should be set to `true` to enable `HTMLWebpackPlugin`.
2. `SOURCEGRAPH_API_URL` is used as a proxied API url. By default it points to the [https://k8s.sgdev.org](https://k8s.sgdev.org).

It's possible to overwrite these variables by creating `sg.config.overwrite.yaml` in the root folder and adjusting the `env` section of the relevant command.

### Development server

```sh
sg run web-standalone
```

For enterprise version:

```sh
sg run enterprise-web-standalone
```

### Production server

```sh
sg run web-standalone-prod
```

For enterprise version:

```sh
sg run enterprise-web-standalone-prod
```

Web app should be available at `http://${SOURCEGRAPH_HTTPS_DOMAIN}:${SOURCEGRAPH_HTTPS_PORT}` (note the `http` not `https`). Build artifacts will be served from `<rootRepoPath>/ui/assets`.

### API proxy

In both environments, server proxies API requests to `SOURCEGRAPH_API_URL` provided as the `.env` variable.
To avoid the `CSRF token is invalid` error CSRF token is retrieved from the `SOURCEGRAPH_API_URL` before the server starts.
Then this value is used for every subsequent request to the API.

### esbuild (experimental)

See https://docs.sourcegraph.com/dev/background-information/web/build#esbuild.
