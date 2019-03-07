# Configuring Sourcegraph with TLS encryption (HTTPS / SSL support) using a self-signed certificate

> NOTE: This tutorial supports Linux only. Windows 10 support is coming soon.

> NOTE: This guide does not yet cover using a Certificate Authority (CA) for issuing certificates which can be trusted by installing the CA on your local machine. This is coming soon too.

> NOTE: Self-signed certificates are ok when initially sharing Sourcegraph with your team internally, but we recommend acquiring a valid (and trusted) certificate through your infrastructure team/person as soon as possible. The same steps below apply for the NGINX configuration section.

> NOTE: Sourcegraph is working on Terraform examples for AWS and soon Google Cloud Platform (GCP), Azure and DigitalOcean that configure Sourcegraph to use a self-signed certificate from the start ([Secure by default](https://en.wikipedia.org/wiki/Secure_by_default)).

In Sourcegraph 3.0+, [NGINX](https://www.nginx.com/resources/glossary/nginx/) is used as the [reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/) for the Sourcegraph HTTP front-end server, making it responsible for [SSL support and termination](https://docs.nginx.com/nginx/admin-guide/security-controls/terminating-ssl-http/).

![NGINX and Sourcegraph architecture](img/sourcegraph-nginx.svg)
Non-sighted users can view a [text-representation of this diagram](img/sourcegraph-nginx.mermaid)

Running Sourcegraph with the [quickstart docker run command](https://docs.sourcegraph.com/) uses the [default NGINX configuration](https://github.com/sourcegraph/sourcegraph/blob/master/cmd/server/shared/assets/nginx.conf) which presumes local usage and TLS encryption (SSL).

Adding TLS support with a self-signed certificate requires two steps:

1. Installing OpenSSL.
1. Creating a self-signed certificate and key.
1. Modifying the default `nginx.conf` file created when Sourcegraph is first initialized.

## 1. Installing OpenSSL

Installing OpenSSL differs for each operating system and Linux distribution so searching the web for **"installing openssl for ubuntu"** is your best bet.

Confirm the `openssl` is on your path by running:

```shell
openssl version
```

## 2. Creating a self-signed certificate and key

Before we commence creating the key, it's important to state some assumptions:

- You're already run Sourcegraph using the `docker run` command from the [Sourcegraph documentation](/../../index.md) which means the `nginx.conf` file is at `~/.sourcegraph/config/nginx.cong`.
- Your [`nginx.conf` file is from version 3.0+](https://github.com/sourcegraph/sourcegraph/blob/master/cmd/server/shared/assets/nginx.conf).
- You're ok with the browser reporting the certificate to be invalid*.

NOTE: *Browsers are now very strict and do not allow certificates with invalid certificates or un-trusted CA's to be trusted/ignored permanently. This is for the most part a good thing. If you haven't already, reach out to your infrastructure team/person to begin the process of acquiring a valid (and trusted) certificate sooner rather than later.

Now that OpenSSL is installed, we can generate the self-signed cert and key:

> NOTE: TODO(ryan): Change the below `openssl` command to have values for all required parameters so the cert and key are generated without user input. The cert won't be valid in any case so it does not matter if the hostname is hardcoded to be `sourcegraph` for example.

```shell
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ~/.sourcegraph/config/sourcegraph.key -out ~/.sourcegraph/config/sourcegraph.crt
```

## 3. Configure NGINX

WIP.
