+++
title = "OAuth2 authentication"
+++

This document uses terminology from
[RFC6749: The OAuth 2.0 Authorization Framework](http://tools.ietf.org/html/rfc6749).

* "Client" is an API client application, including a Sourcegraph
  server acting as a client to another Sourcegraph server.
* "Authorization server" is the Sourcegraph server where the client or
  user registration lives and which grants access tokens.
* "Resource owner" is the user.


# Authentication

Communication with Sourcegraph servers is authenticated with OAuth2.

There are two types of access tokens:

* client access tokens, which authenticate the client (but not a
  user); and
* resource owner access tokens, which authenticate a user *and* the
  client that originated the request.

## Client authentication

Every Sourcegraph instance is authenticated by a public/private
keypair called the "Sourcegraph identity key" (or "ID key" for short),
which is stored at `$SGPATH/id.pem`. This keypair is generated
automatically at server startup if it does not exist.

An instance is identified by its public key fingerprint (hereafter its
"ID"). This is also its OAuth2 client ID. Running `src meta config` or
fetching the HTTP path `/.well-known/sourcegraph` (e.g.,
http://localhost:3000/.well-known/sourcegraph) returns the ID (in the
IDKey field).

Two kinds of authorization grants for client authentication are
derived from the ID key: simple shared secret access tokens and JWT
bearer access tokens. The simple client ID and secret scheme is also
supported.

### Shared secret

If an API client has access to a server's ID key (including the
private key), then the client can generate shared secret access tokens
locally. This avoids the overhead of server-client communication to
grant an access token at no loss to security (since the client already
has the server's private key).

The Web app, HTTP API, and worker use shared secret access tokens to
communicate with the gRPC server. Because these clients currently
always run in the same process as the gRPC server, they are guaranteed
to have access to the server's ID key.

As mentioned in [Security.md](./Security.md), the Web app, HTTP API,
and worker are mere API clients of the server and may only access the
server's resources through the public gRPC API. The use of shared
secret access tokens allows the server to restrict certain operations,
such as account creation and password resets, to being performed
through an even more limited interface (such as the Web app).

Package sharedsecret provides an `oauth2.TokenSource` for generating
shared secret access tokens given an ID key.


### JWT bearer token

External clients must register with the server they wish to
access. This is akin to creating a new GitHub or Google Cloud API
client application. Instead of having the server generate the client
ID and secret, clients may provide a public key in JWKS form during
[client registration](http://openid.net/specs/openid-connect-registration-1_0.html#ClientMetadata),
which is used to authenticate the client in the future. Sourcegraph
clients typically use their ID key's public key here.

After they have registered their public key with the server, clients
may present the server with a signed
[JWT bearer token](https://tools.ietf.org/html/draft-ietf-oauth-jwt-bearer-12)
to obtain an access token.

The `(*IDKey).TokenSource` method produces the JWT bearer token. This
OAuth2 token source is used in the server, not the client. This means
it is only used for server-to-server communication, where one of the
servers acts as the OAuth2 client and the other as the OAuth2
authorization server.

## Client ID and secret

If an external client doesn't provide a JWKS public key during client
registration, then the server generates and returns a client
secret.This is the simplest form of client authentication (used by
GitHub's API, etc.).

Sourcegraph-to-Sourcegraph communication uses public keys and JWT
bearer tokens (not server-generated client IDs and secrets) for
authentication for a few reasons:

* Using the public key fingerprint as the client's ID means that *it*
  determines its own client ID and can use the same client ID
  everywhere (as long as it can back up that claim by presenting JWTs
  signed with the corresponding public key, of course). This provides
  a unique, authenticated identity for each Sourcegraph
  instance. Using externally generated client IDs would not provide
  this.
* When using client secrets, the client must store the secret for each
  server it is registered with. This introduces additional
  complexity. Using JWTs means the client only has to store one
  private key that can authenticate it to every server it has
  registered with. (From an ops POV, this means we just need to
  distribute and secure one key, not an arbitrary and growing number
  of secrets.)


## User (resource owner) authentication

A user account lives on the Sourcegraph server where the account was
originally created (the "authorization server", in OAuth2
terminology), typically Sourcegraph.com. Users may log into the
authorization server directly (e.g., via HTML forms), and session
credentials there are stored in cookies.

Users also can authenticate their identity to Sourcegraph clients
(such as their own Sourcegraph instance) using OAuth2. After the
client receives the authorization code for the user from the
authorization server, it authenticates using its client credentials to
obtain the user's access token.

An access token encodes the user's UID and the client ID that the
access token was generated for.


# ACLs and permissions

Currently, only simple ACLs are defined. Certain API methods, such as
those necessary to log in, are always accessible to everyone. All
other API methods may be restricted to only logged-in users.


# Demo configuration

To set up an environment that mimics the situation where there's a
mothership and standalone Sourcegraph instances, follow these
steps. The end result is that you can register user accounts, etc., on
your demo mothership, and you can authenticate (via OAuth2) via the
mothership to log into your demo standalone instance. This lets you
test features that involve user federation or sending data to or
authenticating via the mothership.

You'll need 2 Sourcegraph instances: the mothership and the local
instance. To make OAuth2 work, they need to be on separate domains. So, first:

```bash
sudo sh -c 'echo 127.0.0.1 demo-mothership >> /etc/hosts'
```

Now run the demo mothership Sourcegraph. Note that it's running on an
entirely separate port, host, and SGPATH--it is as though it's running
on an entirely separate machine (which is what we want to mimic).

```bash
# the mothership
SGPATH=/tmp/mothership make serve-mothership-dev
```

Now run the standalone instance (make sure the demo mothership is
still running):

```
HTTP_DISCOVERY_INSECURE=t make serve-dev SERVEFLAGS='--fed.root-url=http://demo-mothership:13000'
```

Go to http://localhost:3000 to view your standalone instance. Click
Sign In in the top right, and you'll be asked to log in. If you
haven't created an account on the demo mothership yet, create
one.

Next, it asks you to register your new Sourcegraph server. Give it a
name and click Continue.

Then it'll ask you:

```
Authorize My OAuth2 client?
The application at localhost:3000 requests:

Your public user profile (username and company)
You are logged in as sqs. Only proceed if you trust this application.
```

Click the Authorize button, and you are now logged into the standalone
instance with your mothership user account, using an OAuth2 access
token from the mothership that authenticates you to the mothership.

The app session cookie encodes your mothership's OAuth2 access
token. The standalone app and gRPC server treat it as an opaque
value. They pass it along to the mothership untouched to authenticate
you.

In any code path on the mothership that's authenticated using an
actor's OAuth2 access token, the IDKey is also available on the
actor's ClientID field. This can be obtained using
`auth.ActorFromContext(ctx)`. Note that this will not work in app
(frontend app) contexts because the originator of the action there is
a mothership user, not via a standalone instance.


# TODOs

* OAuth2 clients send their shared-secret-derived access tokens to the
  mothership. This is unnecessary and would be good to avoid. We
  assume the mothership is trusted, so it's not a security flaw, but
  it's good to be strict.
