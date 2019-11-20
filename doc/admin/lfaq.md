# Administration Less-Frequently-Asked Questions

## How do I set up redirect URLs in Sourcegraph?

Sometimes URLs in Sourcegraph may change. For example, if an external service configuration is
updated to use a different `repositoryPathPattern`, this will change the repository URLs on
Sourcegraph. Users may wish to preserve links to the old URLs, and this requires adding redirects.

We recommend configuring redirects in a reverse proxy. If you are running Sourcegraph as a single
Docker image, you can deploy a reverse proxy such as [Caddy](https://caddyserver.com/) or
[NGINX](https://www.nginx.com) in front of it. Refer to the
[Caddy](https://github.com/caddyserver/caddy/wiki/v2:-Documentation#rewrite) or
[NGINX](https://www.nginx.com/blog/creating-nginx-rewrite-rules/) documentation for URL rewrites.

If you are running Sourcegraph as a Kubernetes cluster, you have two additional options:

1. If you are using [NGINX
   ingress](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/configure.md#ingress-controller-recommended)
   (`kubectl get ingress | grep sourcegraph-frontend`), modify
   [`sourcegraph-frontend.Ingress.yaml`](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/frontend/sourcegraph-frontend.Ingress.yaml)
   by [adding a rewrite rule](https://kubernetes.github.io/ingress-nginx/examples/rewrite/).
1. If you are using the [NGINX
   service](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/configure.md#nginx-service),
   modify
   [`nginx.ConfigMap.yaml`](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/nginx-svc/nginx.ConfigMap.yaml).
