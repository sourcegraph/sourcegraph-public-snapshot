# Adding SSL (HTTPS) to Sourcegraph with a self-signed certificate

This is for Sourcegraph instances that need a self-signed certificate because they don't yet have certificate from a [globally trusted Certificate Authority (CA) provider](https://en.wikipedia.org/wiki/Certificate_authority#Providers). It works for local and remote (cloud) instances and also shows how to make the self-signed certificate trusted by the browser.

Configuring NGINX to support SSL requires:

1. Installing mkcert.
1. Creating the self-signed certificate.
1. Configuring NGINX for SSL.
1. Changing the `docker run` command to listen on port `443` (for SSL).
1. Configuring your OS so the self-signed certificate is browser trusted

<!-- TODO(ryan): Not sure this is necessary
> NOTE: [Terraform](https://www.terraform.io/intro/index.html) plans are being developed for [AWS](https://github.com/sourcegraph/deploy-sourcegraph-aws), GCP, Azure and DigitalOcean which will be pre-configured to support HTTPS ([Secure by default](https://en.wikipedia.org/wiki/Secure_by_default)). 
-->

> NOTE: See the [Sourcegraph NGINX configuration page](nginx.conf) which has [SSL recommendations for production](nginx.md##nginx-ssl-https-configuration).

## 1. Installing mkcert

The [OpenSSL](https://www.openssl.org/) [CLI](https://wiki.openssl.org/index.php/Command_Line_Utilities) can be used for creating self-signed certificates but its API is challenging unless you're well versed in SSL.

A better alternative is [mkcert](https://github.com/FiloSottile/mkcert#mkcert), an abstraction over OpenSSL written by [Filippo Valsorda](https://github.com/FiloSottile), a cryptographer working at Google on the Go team.

> NOTE: The following commands are to be run on the Docker host, **not** inside the Sourcegraph Docker container.

To set up mkcert for issuing certificates:

1. [Install mkcert for your OS](https://github.com/FiloSottile/mkcert#installation)
1. Create the root [CA](https://en.wikipedia.org/wiki/Certificate_authority) by running:

```shell
mkcert -install
```

## 2. Creating the self-signed certificate

Now that the root CA has been created, mkcert can issue a self-signed certificate (`sourcegraph.crt`) and key (`sourcegraph.key`).

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

Change the [default `~/.sourcegraph/config/nginx.conf`](https://github.com/sourcegraph/sourcegraph/blob/master/cmd/server/shared/assets/nginx.conf) by:

**1.** Replacing `listen 7080;` with `listen 7080 ssl;`.

**2.** Adding the following two lines below the `listen 7080 ssl;` statement.

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

Validate that Sourcegraph is available over https by opening a browser at `https://$HOSTNAME_OR_IP`.

## 5. Configuring your OS so the self-signed certificate is browser trusted

### Sourcegraph hosted locally

You shouldn't see an **Invalid Certificate** warning (except maybe [Firefox](https://github.com/FiloSottile/mkcert#installation)) because `mkcert` created and installed the root CA.

### Sourcegraph hosted externally

You'll be get an **Invalid Certificate** warning because the root CA created on the Sourcegraph machine is not installed/trusted locally.

To install the CA so it is trusted by your local machine's OS (and as a result, your browser):

1. [Install mkcert locally](https://github.com/FiloSottile/mkcert#installation)
1. On the Sourcegraph host, run `mkcert -CAROOT` to get the path to the root CA files (probably `/root/.local/share/mkcert`)
1. Download `rootCA-key.pem` and `rootCA.pem` from the Sourcegraph host to a directory on your local machine, e.g. `~/.mkcert`.
1. Install the root CA by running:

```shell
CAROOT=~/.mkcert mkcert -install
```

**TODO(ryan): Change below instructions to put the root CA files in mkcert's default location for the OS.**
**TODO(ryan): Sample code for downloading from cloud VM to local mkcert default directory.**

Open your browser again at `https://$HOSTNAME_OR_IP` and this time, your certificate should be validated.

## Next steps

The [NGINX SSL Termination](https://docs.nginx.com/nginx/admin-guide/security-controls/terminating-ssl-http/) guide and [Configuring HTTPS Servers](https://nginx.org/en/docs/http/configuring_https_servers.html) contain additional

<!-- 
# Tar gzip root CA files
cd $(mkcert -CAROOT) && tar -czf /home/ubuntu/root-ca.tar.gz root* && chown ubuntu:ubuntu root-ca.tar.gz && cd -


-->
