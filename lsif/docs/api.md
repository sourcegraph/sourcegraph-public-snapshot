# LSIF server endpoints

The LSIF server endpoints are documented as an [OpenAPI v3](https://swagger.io/docs/specification/about/) document [api.yaml](./api.yaml). This document can be viewed locally via docker by running the following command from this directory (or a parent directory if the host path supplied to `-v` changes accordingly).

```bash
docker run \
  -e SWAGGER_JSON=/data/api.yaml \
  -p 8080:8080 \
  -v `pwd`:/data \
  swaggerapi/swagger-ui
```

The OpenAPI document assumes that the LSIF server is running locally on port 3186 in order to make sample requests.

This server should **not** be directly accessible outside of development environments. The endpoints of this server are not authenticated and relies on the Sourcegraph frontend to proxy requests via the HTTP or GraphQL server.
