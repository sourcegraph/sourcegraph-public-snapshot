# Web Application

## Local development

### Configuration

1. Duplicate `client/web/.env.example` as `client/web/.env`.
2. Make sure that `WEBPACK_SERVE_INDEX` is set to `true` in the env file.
3. Make sure that `SOURCEGRAPH_API_URL` points to the accessible API url in the env file.

### Development server

```sh
yarn serve:dev
```

### Production server

```sh
ENTERPRISE=1 NODE_ENV=production DISABLE_TYPECHECKING=true yarn run build
yarn serve:prod
```

Web app should be available at `http://${SOURCEGRAPH_HTTPS_DOMAIN}:${SOURCEGRAPH_HTTPS_PORT}`.
Build artifacts will be served from `<rootRepoPath>/ui/assets`.

### API proxy

In both environments server proxies API requests to `SOURCEGRAPH_API_URL` provided in the `.env` file.
To avoid the `CSRF token is invalid` error CSRF token is retrieved from the `SOURCEGRAPH_API_URL` before the server starts.
Then this value is used for every subsequent request to the API.
