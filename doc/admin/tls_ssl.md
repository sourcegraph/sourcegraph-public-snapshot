# Securing a Sourcegraph instance with TLS/SSL

If you intend to make your Sourcegraph instance accessible on the Internet or another untrusted network, you should use TLS so that all traffic will be served over HTTPS.

See "[nginx HTTP server settings](nginx.md)" for more information.

## Using your own TLS certificate

### Single-server Sourcegraph deployments

For single-server Docker image deployments, add the following lines to your site configuration. The TLS certificate and private key must be specified as PEM-encoded strings.

> Tip: Use [jq](https://stedolan.github.io/jq/) with the command `jq -R --slurp < /path/to/my/cert-or-key.pem` to obtain the JSON-stringified contents of each PEM file.

```json
{
  // ...
  "tlsCert": "-----BEGIN CERTIFICATE-----\nMIIFdTCCBF2gAWiB...",
  "tlsKey": "-----BEGIN RSA PRIVATE KEY-----\nMII...",
  "externalURL": "https://example.com:3443" // Must begin with "https"; replace with the public IP or hostname of your machine
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
  sourcegraph/server:2.13.2
```

If you are running on cloud infrastructure, you will likely need to add an ingress rule to make port 30443 accessible to the Internet.
