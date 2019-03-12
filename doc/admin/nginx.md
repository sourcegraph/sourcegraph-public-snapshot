# Sourcegraph NGINX HTTP and HTTPS/SSL configuration

In Sourcegraph 3.0+, [NGINX](https://www.nginx.com/resources/glossary/nginx/) acts as a [reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/) for the Sourcegraph front-end server, meaning NGINX proxies external HTTP (and [HTTPS](#nginx-ssl-https-configuration)) requests to the Sourcegraph front-end.

![NGINX and Sourcegraph architecture](img/sourcegraph-nginx.svg)

**Note**: Non-sighted users can view a [text-representation of the NGINX and Sourcegraph front-end HTTP flow diagram](img/sourcegraph-nginx.mermaid) in [Mermaid.js format](https://mermaidjs.github.io/).

As the initial NGINX configuration is suited for local/internal usage, this page provides additional code recommended for external/production deployment.

## NGINX for Sourcegraph single instance (Docker)

<!-- TODO(ryan): Change heading to ## The default `nginx.conf` file and how to extend/override default configuration and add section
on how to extend NGINX configuration without (in most cases), editing the `nginx.conf` file. -->

When Sourcegraph is first run, it will create a [default `nginx.conf`](https://github.com/sourcegraph/sourcegraph/blob/master/cmd/server/shared/assets/nginx.conf) at:

- `/etc/sourcegraph/nginx.conf` inside the container, and
- `~/.sourcegraph/config/nginx.conf` on the Docker/Sourcegraph host ((presuming you're using the [quickstart `docker run` command](../index.md#quickstart))).

## NGINX for Sourcegraph Cluster (Kubernetes)

We use [ingress-nginx](https://kubernetes.github.io/ingress-nginx/) for Sourcegraph Cluster. Refer to the [deploy-sourcegraph Configuration](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/configure.md) documentation for more information.

## NGINX for other Sourcegraph clusters (e.g. pure-Docker)

The pure-Docker deployment reference ([deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker)) aims to be minimal and not tied to any specific deployment method, so we don't bundle NGINX in there. You can use any reverse proxy to provide HTTPS for your Sourcegraph instance.

We suggest using [the official NGINX docker images](https://hub.docker.com/_/nginx) and following their instructions for [securing HTTP traffic with a proxied server](https://docs.nginx.com/nginx/admin-guide/security-controls/securing-http-traffic-upstream/).

Finally, you should configure Sourcegraph's `externalURL` in the [critical configuration](config/critical_config.md) (and restart the frontend instances) so that Sourcegraph knows its URL.

## NGINX SSL/HTTPS configuration

- **[Self-signed certificates on NGINX](ssl_https_self_signed_cert_nginx.md)**: Great for new instances that don't yet have a browser trusted certificate.
- **[Let's Encrypt (Certbot) on NGINX](https://certbot.eff.org/)**: NGINX supported certificate management tool for programmatically obtaining a browser-trusted certificate.

## Redirect to external HTTPS URL

The URL that clients should use to access Sourcegraph is defined in the `externalURL` property in [critical configuration](config/critical_config.md). To enforce that clients access Sourcegraph via this URL (and not some other URL, such as an IP address or other non-`https` URL), add the following to `nginx.conf` (replacing `https://sourcegraph.example.com` with your external URL):

``` nginx
# Redirect non-HTTPS traffic to HTTPS.
server {
    listen 80;
    server_name _;

    location / {
        return 301 https://yourdomain.com$request_uri;
    }
}
```

## HTTP Strict Transport Security

[HTTP Strict Transport Security](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) instructs web clients to only communicate with the server over HTTPS. To configure it, add the following to `nginx.conf` (in the `server` block):

``` nginx
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

See [`add_header` documentation](https://nginx.org/en/docs/http/ngx_http_headers_module.html#add_header) and "[Configuring HSTS in nginx](https://www.nginx.com/blog/http-strict-transport-security-hsts-and-nginx/)" for more details.
