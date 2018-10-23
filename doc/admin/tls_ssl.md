# Securing a Sourcegraph instance with TLS/SSL

If you intend to make your Sourcegraph instance accessible on the Internet or another untrusted network, you should use TLS so that all traffic will be served over HTTPS.

## Let's Encrypt

Sourcegraph will use [Let's Encrypt](https://letsencrypt.org/) by default if the following conditions are met:

- Your `appURL` site configuration option begins with `https://...`.
- The host is reachable on both ports `80` and `443` (see [note](#port-80-must-be-accessible) below).
- You have not configured manual TLS certificates as described below.
- You have not configured `tls.letsencrypt` to `off`. (Defaults to `auto`)

Once you have HTTPS, working we suggest configuring `httpToHttpsRedirect` to `true` to prevent users browsing Sourcegraph via plaintext HTTP.

### Port 80 must be accessible

[Let's Encrypt requires that port `80` be reachable in order to prove that you own your domain](https://letsencrypt.readthedocs.io/en/latest/challenges.html#http-01-challenge). If port `80` is unreachable, HTTPS will fail with errors such as the following:

```bash
http: TLS handshake error from 10.240.0.17:11486: acme/autocert: unable to authorize "example.com"; challenge "tls-alpn-01" failed with error: acme: authorization error for example.com: 403 urn:acme:error:unauthorized: Cannot negotiate ALPN protocol "acme-tls/1" for tls-alpn-01 challenge; challenge "http-01" failed with error: acme: authorization error for example.com: 403 urn:acme:error:unauthorized: Invalid response from http://example.com/.well-known/acme-challenge/gHyMIbdfCVRvnz0FUJuezDsDJYD7flbVBzr348MrfLg: "<!DOCTYPE html>\n<!--[if lt IE 7]> <html class=\"no-js ie6 oldie\" lang=\"en-US\"> <![endif]-->\n<!--[if IE 7]>    <html class=\"no-js "

...

http: TLS handshake error from 10.20.3.1:13676: acme/autocert: missing certificate

...

http: TLS handshake error from 10.240.0.16:41012: 429 urn:acme:error:rateLimited: Error creating new authz :: too many failed authorizations recently: see https://letsencrypt.org/docs/rate-limits/
```

---

## Using your own TLS certificate

### Single-server Sourcegraph deployments

For single-server Docker image deployments, add the following lines to your site configuration. The TLS certificate and private key must be specified as PEM-encoded strings.

> Tip: Use [jq](https://stedolan.github.io/jq/) with the command `jq -R --slurp < /path/to/my/cert-or-key.pem` to obtain the JSON-stringified contents of each PEM file.

```json
{
  // ...
  "tlsCert": "-----BEGIN CERTIFICATE-----\nMIIFdTCCBF2gAWiB...",
  "tlsKey": "-----BEGIN RSA PRIVATE KEY-----\nMII...",
  "appURL": "https://example.com:3443" // Must begin with "https"; replace with the public IP or hostname of your machine
  // ...
}
```

Next, restart your Sourcegraph instance using the same `docker run` [command](install/docker/index.md), but map the host port to the container HTTPS port 7443 (not the HTTP port 7080). In this example, the host port 443 (HTTPS) is mapped to the container's HTTPS port 7443.

```shell
docker run \
  --publish 443:7443 --rm \
  --volume ~/.sourcegraph/config:/etc/sourcegraph \
  --volume ~/.sourcegraph/data:/var/opt/sourcegraph \
  --volume /var/run/docker.sock:/var/run/docker.sock \
  sourcegraph/server:2.12.2
```

If you are running on cloud infrastructure, you will likely need to add an ingress rule to make port 30443 accessible to the Internet.
