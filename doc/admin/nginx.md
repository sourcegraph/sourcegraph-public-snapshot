# Sourcegraph NGINX HTTP and HTTPS/SSL configuration

> NOTE: This a Sourcegraph 3.0+ feature.

In Sourcegraph 3.0+, [NGINX](https://www.nginx.com/resources/glossary/nginx/) acts as a [reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/) for the Sourcegraph front-end server, meaning NGINX proxies external HTTP (and [HTTPS](#nginx-ssl-https-configuration)) requests to the Sourcegraph front-end.

![NGINX and Sourcegraph architecture](img/sourcegraph-nginx.svg)

**Note**: Non-sighted users can view a [text-representation of this diagram](sourcegraph-nginx-mermaid.md).

## NGINX for Sourcegraph single instance (Docker)

<!-- TODO(ryan): Change heading to ## The default `nginx.conf` file and how to extend/override default configuration and add section
on how to extend NGINX configuration without (in most cases), editing the `nginx.conf` file. -->

The first time Sourcegraph is run, it will create an [`nginx.conf`](https://github.com/sourcegraph/sourcegraph/blob/master/cmd/server/shared/assets/nginx.conf) file at:

- `~/.sourcegraph/config/nginx.conf` on the Docker/Sourcegraph host (presuming you're using the [quickstart `docker run` command](../index.md#quickstart)))
- `/etc/sourcegraph/nginx.conf` inside the container

[SSL support requires manual editing](#nginx-ssl-https-configuration) of the NGINX configuration file if using the [quickstart docker run command](../index.md#quickstart) as it presumes local or internal usage.

## NGINX for Sourcegraph Cluster (Kubernetes)

We use the [ingress-nginx](https://kubernetes.github.io/ingress-nginx/) for Sourcegraph Cluster running on Kubernetes. Refer to the [deploy-sourcegraph Configuration](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/configure.md) documentation for more information.

## NGINX for other Sourcegraph clusters (e.g. pure-Docker)

NGINX is not included in the ([pure-Docker deployment](https://github.com/sourcegraph/deploy-sourcegraph-docker) as it's designed to be minimal and not tied to any specific reverse proxy.

If NGINX is your preferred reverse proxy, we suggest using [the official NGINX docker images](https://hub.docker.com/_/nginx) and following their instructions for [securing HTTP traffic with a proxied server](https://docs.nginx.com/nginx/admin-guide/security-controls/securing-http-traffic-upstream/).

## NGINX SSL/HTTPS configuration

### If you have a valid SSL certificate

**1.** Copy your SSL certificate and key to `~/.sourcegraph/config` (where the `nginx.conf` file is).

**2.** Edit `nginx.conf`, replacing `listen 7080;` with `listen 7080 ssl;`, then add the following two lines below the `listen 7080 ssl;` statement (names of cert and key don't matter).

```nginx
ssl_certificate         sourcegraph.crt;
ssl_certificate_key     sourcegraph.key;
```

The `nginx.conf` should now look like:

```nginx
...
http {
    ...
    server {
       ...
        listen 7080 ssl;
        ssl_certificate         sourcegraph.crt;
        ssl_certificate_key     sourcegraph.key;
        ...
    }
}
```

### If you need an SSL certificate

There are a few options:

**[1. Generate a self-signed certificate](ssl_https_self_signed_cert_nginx.md)**<br />
For instances that don't yet have certificate from a [globally trusted Certificate Authority (CA) provider](https://en.wikipedia.org/wiki/Certificate_authority#Providers).

**[2. Generate a browser trusted certificate using Let's Encrypt (Certbot)](https://certbot.eff.org/)**<br />
NGINX supported certificate management tool for programmatically obtaining a globally browser-trusted certificate.

**3. Proxy as a service**<br />
Services such as [Cloudflare](https://www.cloudflare.com/ssl/) can handle the SSL connection from the browser/client, proxying requests to your Sourcegraph instance.

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

## Additional NGINX SSL configuration

See the [NGINX SSL Termination](https://docs.nginx.com/nginx/admin-guide/security-controls/terminating-ssl-http/) guide and [Configuring HTTPS Servers](https://nginx.org/en/docs/http/configuring_https_servers.html).

## Next steps

You should configure Sourcegraph's `externalURL` in the [critical configuration](config/critical_config.md) (and restart the frontend instances) so that Sourcegraph knows its URL.
