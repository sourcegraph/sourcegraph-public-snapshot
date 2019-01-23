# nginx HTTP server settings

Sourcegraph uses [nginx](https://nginx.org/en/) to proxy HTTP traffic between clients and the Sourcegraph HTTP server. It ships with a default nginx configuration that is intended for local/internal network usage.

On initial startup Sourcegraph will generate an `nginx.conf` which you can modify. It is located at `/etc/sourcegraph/nginx.conf` in the container. So if you use the quick start docker run command it will be at `~/.sourcegraph/config/nginx.conf`. (due to the docker flag `--volume ~/.sourcegraph/config:/etc/sourcegraph`).

## TLS / HTTPS

If you have a TLS certificate and key to use for Sourcegraph, you can setup nginx to terminate TLS. First copy your TLS certificate and key into the same directory as `nginx.conf`:

```shell
cp sourcegraph.example.com.crt ~/.sourcegraph/config/
cp sourcegraph.example.com.key ~/.sourcegraph/config/
```

Then you can configure nginx to listen with ssl using the above certificates:

```nginx
error_log stderr;
pid /var/run/nginx.pid;

events {
}

http {
  server_tokens off;

  # We can upload large extensions
  client_max_body_size 150M;

  # Don't timeout websockets quickly. Default is 60s. This is the timeout
  # between reads/writes, not the full session timeout.
  proxy_send_timeout 1h;
  proxy_read_timeout 1h;

  access_log off;
  upstream backend {
    server localhost:8080;
  }

  server {
    listen 7080 ssl;

    # Make sure the .crt and .key files exist in the same directory as nginx.conf!
    ssl_certificate     YOUR_CERTIFICATE.crt;
    ssl_certificate_key YOUR_KEY.key;

    location / {
      proxy_pass http://backend;
      proxy_set_header Host $http_host;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
    }
  }
}
```

Next, <kbd>Ctrl-C</kbd> your Sourcegraph instance and re-run the [`docker run` command](install/docker/index.md) and refresh the page.

To run on port 443, specify `--publish 443:7080` instead of `--publish 7080:7080` in the `docker run` command.

See [NGINX SSL Termination](https://docs.nginx.com/nginx/admin-guide/security-controls/terminating-ssl-http/) guide and [Configuring HTTPS Servers](https://nginx.org/en/docs/http/configuring_https_servers.html) for more information.

## Let's Encrypt

[Let's Encrypt](https://letsencrypt.org) automatically provisions TLS certificates so that your server is accessible via HTTPS. You can configure it with nginx using EFF's [Certbot](https://certbot.eff.org/), which has instructions for most common setups:

- [Using Let's Encrypt with nginx on Ubuntu 18.04](https://certbot.eff.org/lets-encrypt/ubuntubionic-nginx)
- [Using Let's Encrypt with nginx on Ubuntu 16.04](https://certbot.eff.org/lets-encrypt/ubuntuxenial-nginx)
- [Using Let's Encrypt with nginx on Debian 9](https://certbot.eff.org/lets-encrypt/debianstretch-nginx)
- [Using Let's Encrypt with nginx on CentOS/RHEL 7](https://certbot.eff.org/lets-encrypt/centosrhel7-nginx)
- [Using Let's Encrypt with nginx on macOS](https://certbot.eff.org/lets-encrypt/osx-nginx)

Use the dropdown menus on the Certbot site to find instructions for other setups.

## Redirect to external HTTPS URL

The URL that clients should use to access Sourcegraph is called the [`externalURL`](site_config/all.md#externalurl-string) in site configuration. To enforce that clients access Sourcegraph via this URL (and not some other URL, such as an IP address or other non-`https` URL), add the following to `nginx.conf` (replacing `https://sourcegraph.example.com` with your external URL):

``` nginx
# Redirect non-HTTPS traffic to HTTPS.
server {
    listen 80;
    server_name _;

    # Uncomment this block if you are using Let's Encrypt (otherwise it will be unable to
    # communicate with your server to generate the TLS certificate).
    #
    # location /.well-known/acme-challenge/ {
    #    try_files $uri =404;
    # }

    location / {
        # REPLACE https://sourcegraph.example.com with your external URL:
        return 301 https://sourcegraph.example.com$request_uri;
    }
}
```

## HTTP Strict Transport Security

[HTTP Strict Transport Security](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) instructs web clients to only communicate with the server over HTTPS. To configure it, add the following to `nginx.conf` (in the `server` block):

``` nginx
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

See [`add_header` documentation](https://nginx.org/en/docs/http/ngx_http_headers_module.html#add_header) and "[Configuring HSTS in nginx](https://www.nginx.com/blog/http-strict-transport-security-hsts-and-nginx/)" for more details.

## nginx for Sourcegraph Cluster (Kubernetes)

We use [ingress-nginx](https://kubernetes.github.io/ingress-nginx/) for Sourcegraph Cluster. Refer to the [deploy-sourcegraph Configuration](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/configure.md) documentation for more information.

## nginx for other Sourcegraph clusters (e.g. pure-Docker)

The pure-Docker deployment reference ([deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker)) aims to be minimal and not tied to any specific deployment method, so we don't bundle nginx in there. You can use any reverse proxy to provide HTTPS for your Sourcegraph instance.

We suggest using [the official nginx docker images](https://hub.docker.com/_/nginx) and following their instructions for [securing HTTP traffic with a proxied server](https://docs.nginx.com/nginx/admin-guide/security-controls/securing-http-traffic-upstream/).

Lastly, you should configure Sourcegraph's [`externalURL`](site_config/all.md#externalurl-string) in the [management console](management_console.md) (and restart the frontend instances) so that Sourcegraph knows its URL.
