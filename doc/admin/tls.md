---
ignoreDisconnectedPageCheck: true
---

# Adding HTTPS/SSL to Sourcegraph using a self-signed certificate

> NOTE: All commands are to be run on the Docker host, **not** in the Sourcegraph Docker container.

In Sourcegraph 3.0+, [NGINX](https://www.nginx.com/resources/glossary/nginx/) acts as a [reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/) for the Sourcegraph front-end server, meaning NGINX proxies external HTTP (and HTTPS) requests to the Sourcegraph front-end.

![NGINX and Sourcegraph architecture](img/sourcegraph-nginx.svg)
<p class="text-center small">Note: Non-sighted users can view a [text-representation of this diagram](img/sourcegraph-nginx.mermaid)</p>

The [quickstart docker run command](https://docs.sourcegraph.com/#quickstart-guide) assumes it's for local and/or internal usage which is why SSL is not pre-configured.

Configuring NGINX to support SSL requires:

1. Installing mkcert.
1. Creating the self-signed certificate.
1. Adding SSL support to NGINX.
1. Changing the `docker run` command to listen on port `443` (for SSL).

> NOTE: [Terraform](https://www.terraform.io/intro/index.html) plans are being developed for [AWS](https://github.com/sourcegraph/deploy-sourcegraph-aws), GCP, Azure and DigitalOcean which will be pre-configured to support HTTPS ([Secure by default](https://en.wikipedia.org/wiki/Secure_by_default)).

## 1. Installing mkcert

The [OpenSSL](https://www.openssl.org/) [CLI](https://wiki.openssl.org/index.php/Command_Line_Utilities) can be used for creating self-signed certificates but its API is challenging unless you're well versed in SSL.

A better alternative is [mkcert](https://github.com/FiloSottile/mkcert#mkcert), an abstraction over OpenSSL written by [Filippo Valsorda](https://github.com/FiloSottile), a cryptographer working at Google on the Go team.

To set up mkcert for issuing certificates:

- [Install mkcert for your OS](https://github.com/FiloSottile/mkcert#installation)
- Create the root [CA](https://en.wikipedia.org/wiki/Certificate_authority) by running:

```shell
mkcert -install
```

mkcert is now ready to create self-signed certificates.

## 2. Creating the self-signed certificate

We'll use `mkcert` to create the self-signed certificate (`sourcegraph.crt`) and key (`sourcegraph.key`).

> NOTE: Replace `$HOSTNAME_OR_IP` in the code below with the external hostname or IP address of the Sourcegraph host.

```shell
mkcert \
  -cert-file ~/.sourcegraph/config/sourcegraph.crt \
  -key-file ~/.sourcegraph/config/sourcegraph.key \
  $HOSTNAME_OR_IP
```

Run `ls -la ~/.sourcegraph/config/` and you should see `nginx.conf`, `sourcegraph.crt`, and `sourcegraph.crt` listed together.

<!-- TODO (ryan): Decide if this content is worth keeping
>> NOTE: If you don't want to use `mkcert`, you can create the certificate and key using OpenSSL:
> 
> ```shell
> openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ~/.sourcegraph/config/sourcegraph.key -out ~/.sourcegraph/config/sourcegraph.crt -subj "/CN=localhost"
> ```
-->

## 3. Adding SSL support to NGINX

Open the `~/.sourcegraph/config/nginx.conf` file ([should look like this](https://github.com/sourcegraph/sourcegraph/blob/master/cmd/server/shared/assets/nginx.conf)) and make the following changes:

**1.** Replace `listen 7080;` with `listen 7080 ssl;`

**2.** Then add the following two lines below the `listen 7080 ssl;` statement.

```nginx
ssl_certificate         sourcegraph.crt;
ssl_certificate_key     sourcegraph.key;
```

## 4. Changing the quickstart `docker run` command to listen on port `443` (for SSL)

Now that NGINX is listening on port 443, add `--publish 443:7080` to the `docker run` command::

```shell
docker container run \
  --rm  \
  --publish 7080:7080 \
  --publish 2633:2633 \
  --publish 443:7080 \
  \
  --volume ~/.sourcegraph/config:/etc/sourcegraph  \
  --volume ~/.sourcegraph/data:/var/opt/sourcegraph  \
  sourcegraph/server:3.1.1
```

## Testing Sourcegraph with HTTPS

Validate that Sourcegraph is available over https by opening a browser at `https://$HOSTNAME_OR_IP`.

## If Sourcegraph is on the same machine as your browser, e.g. macOS

You shouldn't see an **Invalid Certificate** warning (except maybe [Firefox](https://github.com/FiloSottile/mkcert#installation)) because `mkcert` created and installed the root CA.

## If Sourcegraph is on a different machine than your browser, e.g. macOS and AWS EC2 instance

You'll be get an **Invalid Certificate** warning because the root CA created on the server is not installed/trusted locally.

### Installing the root CA on your local machine:

Let's remove the **Invalid Certificate** warning by installing the root CA locally:

- On the Docker/Sourcegraph host, run `mkcert -CAROOT` to get the path to the root CA files (probably `/root/.local/share/mkcert`)
- Download `rootCA-key.pem` and `rootCA.pem` to a directory on your local machine, e.g. `~/.mkcert`.
- Then install the root CA by running:

```shell
CAROOT=~/.mkcert mkcert -install
```

Open your browser again at `https://$HOSTNAME_OR_IP` and this time, your certificate should be validated.

<!-- 
# Tar gzip root CA files
cd $(mkcert -CAROOT) && tar -czf /home/ubuntu/root-ca.tar.gz root* && chown ubuntu:ubuntu root-ca.tar.gz && cd -


-->
