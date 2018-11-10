# nginx HTTP server settings

Sourcegraph uses [nginx](https://nginx.org/en/) to proxy HTTP traffic between clients and the Sourcegraph HTTP server. It ships with a default nginx configuration that is intended for local/internal network usage.

## HTTP Strict Transport Security

[HTTP Strict Transport Security](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) instructs web clients to only communicate with the server over HTTPS. To configure it, add the following to `nginx.conf` (in the `server` block):

``` nginx
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

See [`add_header` documentation](https://nginx.org/en/docs/http/ngx_http_headers_module.html#add_header) and "[Configuring HSTS in nginx](https://www.nginx.com/blog/http-strict-transport-security-hsts-and-nginx/)" for more details.
