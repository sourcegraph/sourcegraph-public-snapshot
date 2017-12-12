---
layout: markdown
title: Sourcegraph Server
permalink: docs/preview/docker-images
---

# Sourcegraph Server - Docker images

[_Watch a 60-second GIF screencast of these instructions_](https://images.contentful.com/le3mxztn6yoo/24wwIXOnsIuqc00UAA26WU/9b44e95746b5deda97861e409fe4b751/Great_code_search_60_seconds.gif)

[Sourcegraph Server](https://about.sourcegraph.com/products/server) is available
on Docker Hub:

* [sourcegraph/server (Docker image)](https://hub.docker.com/r/sourcegraph/server/)

This Docker image contains everything necessary to run Sourcegraph Server.

Sourcegraph Server provides fast, powerful code search. For code intelligence,
high availability, and high scalability,
[contact us](mailto:support@sourcegraph.com) for Sourcegraph Enterprise.

## Prerequisites

Docker is required. See the
[official installation docs](https://docs.docker.com/engine/installation/).
(**Note:** Using a native Docker install instead of Docker Toolbox is
recommended in order to use the persisted volumes.)

## Run the Docker image

The following command runs Sourcegraph Server.

```sh
docker run \
 --rm \
 -e ORIGIN_MAP='github.com/!https://github.com/%.git' \
 --publish 3080:80 \
 --volume /tmp/sourcegraph/config:/etc/sourcegraph \
 --volume /tmp/sourcegraph/data:/var/opt/sourcegraph \
 sourcegraph/server:latest
```

**Done!** Sourcegraph is now available on the web at http://localhost:3080.

(If your server has access to the Internet, it's now ready to use with any
public GitHub repository. Try visiting
http://localhost:3080/github.com/sourcegraph/hello.)

Continue to the next section to add your own repositories.

## Add repositories

(To change environment variables of your Sourcegraph Server, kill the container
and rerun it with the new values.)

Set the container's `ORIGIN_MAP` environment variable to tell Sourcegraph how to
add and clone repositories. For example:

* `ORIGIN_MAP=github.com/!https://github.com/%.git` means that GitHub
  repositories are available on Sourcegraph at
  `http(s)://[sourcegraph-hostname]/github.com/foo/bar`
* `ORIGIN_MAP=github.example.com/!git@github.example.com:%.git` is similar to
  the above, but for GitHub Enterprise and using SSH cloning (see the next
  section for how to authenticate cloning)
* `ORIGIN_MAP=gitlab.example.com/!https://gitlab.example.com/%.git` means that a
  repository whose clone URL is `https://gitlab.example.com/foo/bar.git` is
  available on Sourcegraph at
  `http(s)://[sourcegraph-hostname]/gitlab.example.com/foo/bar`

Visiting a repository on Sourcegraph for the first time automatically adds and
clones it according to the configuration. For example, to add a repository
`gitlab.example.com/foo/bar` (which is matched by the last `ORIGIN_MAP` example
value above), visit:

```text
http(s)://[sourcegraph-hostname]/gitlab.example.com/foo/bar
```

You'll see an indicator as it clones the repository. When cloning finishes,
you'll see a file listing. The repository is now immediately searchable at all
revisions.

### Repositories that need HTTP(S) or SSH authentication

If authentication is required to `git clone` the repository clone URLs that
`ORIGIN_MAP` specifies, then you must add credentials to the container's config
volume (as described below for HTTP(S) and SSH) and restart the container.

The following commands demonstrate how to copy the appropriate credentials from
your local machine:

* For HTTP(S) cloning: `cp ~/.netrc /tmp/sourcegraph/config/netrc`
* For SSH cloning: `cp -r ~/.ssh/ /tmp/sourcegraph/config/ssh`

## Where is the data stored?

The Sourcegraph container uses host-mounted volumes to store persistent data:

| Local location            | Container location     | Usage                                           |
| ------------------------- | ---------------------- | ----------------------------------------------- |
| `/srv/sourcegraph/config` | `/etc/sourcegraph`     | For storing the Sourcegraph configuration files |
| `/srv/sourcegraph/data`   | `/var/opt/sourcegraph` | For storing application data                    |

(This assumes that you created the `config` and `data` host volume directories
underneath `/srv/sourcegraph`.)

## Configure Sourcegraph

Sourcegraph is configured via the container's environment variables. You can
also set environment variables in the container file `/etc/sourcegraph/env` by
appending lines of the form `NAME=VALUE`.

See
[Sourcegraph Server documentation](https://about.sourcegraph.com/docs/server/)
for configuration options. **NOTE:** The documentation describes the new JSON
configuration format, but currently the Docker image only accepts the old
environment variable configuration. Generally the environment variable name is
derivable from the JSON field name (e.g., `searchScopes` maps to
`SEARCH_SCOPES`). Contact us for help configuring the container as needed.

## User authentication

### SAML

Create a new SAML app in your SSO provider. Specify the following values in the
SAML configuration for your app in your SSO provider's settings:

* **Single Sign-On URL, Recipient URL, Destination URL:**
  `http(s)://[sourcegraph-hostname]/.auth/saml/acs`
* **Audience URI / SP Entity ID:**
  `http(s)://[sourcegraph-hostname]/.auth/saml/metadata`
* **Attribute statements:** These are the user attributes returned to the
  service provider in the SAML assertion. Some identity providers set these by
  default; others require them to be set explicitly. Ensure that the following
  attributes exist in the SAML assertion:<br> `Email` (required): the user's
  email<br> `Login` (optional): the user's login name (used in @mentions)<br>
  `DisplayName` (optional): the name to display in the nav bar (typically the
  user's first name)<br>

Add the following values to your Docker environment variable file:

```text
SAML_ID_PROVIDER_METADATA_URL=[the metadata URL of your SAML identity provider]
SAML_CERT=[the SAML service provider certificate]
SAML_KEY=[the SAML service provider private key]
```

The SAML identity provider metadata URL should be listed in the documentation of
your SSO provider. The SAML service provider certificate and private key are
values you generate that are used to validate requests from Sourcegraph (the
SAML service provider) to the identity provider. You can generate a key and
certificate with:

```text
openssl req -x509 -newkey rsa:4096 -keyout saml-sp.key -out saml-sp.cert -days 365 -nodes -subj "/CN=myservice.example.com"
```

### OpenID Connect

Create a new OpenID client app in your SSO provider.

Then add the following values to your Docker environment variable file:

```text
SRC_APP_URL=http(s)://[sourcegraph-hostname]
OIDC_OP=[URL to your OpenID Provider]
OIDC_CLIENT_ID=[the OIDC client ID]
OIDC_CLIENT_SECRET=[the OIDC client secret]
OIDC_EMAIL_DOMAIN=[an email hostname for restricting SSO logins; if set, only user accounts with emails assocaited with the domain will be allowed to sign in; this is optional and only recommended if your SSO provider is multi-tenant ]
```

In your SSO provider settings, set the OIDC callback URL of your client
application to `http(s)://[sourcegraph-hostname]/.auth/callback`.

## External database

The image includes self-contained instances of
[PostgreSQL](https://www.postgresql.org/) (for user data) and
[Redis](https://redis.io/) (for session data). Their data is stored in the
container directory `/var/opt/sourcegraph`.

You can use external services instead (e.g., to use your existing
PostgreSQL/Redis databases or to use managed databases from your cloud
provider):

* PostgreSQL: set the standard
  [PG\* environment variables](https://www.postgresql.org/docs/9.4/static/libpq-envars.html),
  such as `PGHOST` and `PGPORT`. You may prefer to use a
  [`PGDATASOURCE` connection string](https://www.postgresql.org/docs/9.4/static/libpq-connect.html#LIBPQ-CONNSTRING).
* Redis: set the environment variable `REDIS=hostname:port`

Even if you've configured both PostgreSQL and Redis to use external services,
you still need to maintain the volume mapping for `/var/opt/sourcegraph` because
it still contains cached repository data (cloned from your code host).
