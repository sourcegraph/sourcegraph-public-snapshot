---
ignoreDisconnectedPageCheck: true
---

# Configure Sourcegraph to support HTTPS with NGINX SSL configuration

> NOTE: This tutorial supports Linux only.

---

> NOTE: Self-signed certificates are ok when sharing Sourcegraph with your team initially, but we recommend acquiring a valid (and trusted) certificate.

---

> NOTE: Sourcegraph is working on [Terraform](https://www.terraform.io/intro/index.html) examples that come pre-configured to support HTTPS ([Secure by default](https://en.wikipedia.org/wiki/Secure_by_default)).

In Sourcegraph 3.0+, [NGINX](https://www.nginx.com/resources/glossary/nginx/) is used as the [reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/) for the Sourcegraph HTTP front-end server, making it responsible for [SSL support and termination](https://docs.nginx.com/nginx/admin-guide/security-controls/terminating-ssl-http/).

![NGINX and Sourcegraph architecture](img/sourcegraph-nginx.svg)
Non-sighted users can view a [text-representation of this diagram](img/sourcegraph-nginx.mermaid)

Running Sourcegraph with the [quickstart docker run command](https://docs.sourcegraph.com/) uses the [default NGINX configuration](https://github.com/sourcegraph/sourcegraph/blob/master/cmd/server/shared/assets/nginx.conf) which presumes local usage and TLS encryption (SSL).

There are four steps required SSL support wth a self-signed certificate:

1. Install [mkcert](https://github.com/FiloSottile/mkcert)
1. Create a self-signed certificate and key.
1. Modify the NGINX for SSL
1. Change `docker run` command to listen on port `443` for SSL.

> NOTE: These commands are to be run on the host the Sourcegraph container is running on, **not** inside the Sourcegraph container.

## Prerequisites

The `docker run` command from the [quickstart guide](https://docs.sourcegraph.com) must be run so the `nginx.conf` file will exist at `~/.sourcegraph/config/nginx.conf`.


## 1. Install mkcert

To create the certificate, we'll use [mkcert](https://github.com/FiloSottile/mkcert#mkcert), an abstraction over OpenSSL with a lovely API written by [Filippo Valsorda](https://github.com/FiloSottile), a cryptographer working at Google on the Go team.

Follow the [installation instructions](https://github.com/FiloSottile/mkcert#installation) for your OS.

Then create the root [Certificate Authority](https://en.wikipedia.org/wiki/Certificate_authority) (CA) by running:

```shell
mkcert -install
```

We're now ready for mkcert to issue the certificate and key

## 2. Creating a self-signed certificate and key

> NOTE: *Self-signed or invalid certificates are now not able to be marked as trusted *within* the browser. Instead, the certificate and the CA that issued it must be installed and trusted by the host operating system.

Now let's use `mkcert` to create the self-signed certificate and key:

```shell
mkcert -cert-file ~/.sourcegraph/config/sourcegraph.crt -key-file ~/.sourcegraph/config/sourcegraph.key localhost
```

> NOTE: If you don't want to use `mkcert`, you can create the certificate and key using OpenSSL:

```shell
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ~/.sourcegraph/config/sourcegraph.key -out ~/.sourcegraph/config/sourcegraph.crt -subj "/CN=localhost"
```

## 3. Modify the NGINX for SSL

Open the `~/.sourcegraph/nginx.conf` file and make the following changes:

**1.** Replace `listen 7080;` with `listen 7080 ssl;`

**2.** Then add the following two lines below the `listen 7080 ssl;` statement.

```nginx
ssl_certificate         sourcegraph.crt;
ssl_certificate_key     sourcegraph.key;
```

## 4. Change `docker run` command to listen on port `443` for SSL

Now that NGINX is listening on port 443, add `--publish 443:7080` to the `docker run` command. It now should resemble this:

```shell
docker container run \
  --publish 7080:7080  \
  --publish: 443:7080 \
  --publish 2633:2633  \
  --rm  \
  --volume ~/.sourcegraph/config:/etc/sourcegraph  \
  --volume ~/.sourcegraph/data:/var/opt/sourcegraph  \
  sourcegraph/server:3.1.1
```

Now open your browser to `https:///localhost` substitute `localhost` for the IP address of the host machine.

> NOTE: It's up to you whether you keep `--publish 7080:7080` as technically you don't need it anymore
