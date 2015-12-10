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
which is stored at `$SGPATH/appdata/global/core.serve/auth/id.pem`. This
keypair is generated automatically at server startup if it does not exist and
can alternatively be read from the environment variable `SRC_ID_KEY_DATA`.

An instance is identified by its public key fingerprint (hereafter its
"ID"). This is also its OAuth2 client ID. Running `src meta config` or
fetching the HTTP path `/.well-known/sourcegraph` (e.g.,
http://localhost:3080/.well-known/sourcegraph) returns the ID (in the
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

As mentioned in [Security.md]({{< relref "dev/Security.md" >}}), the Web app, HTTP API,
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

After they have registered their public key with the server (which
happens automatically on first-run), clients may present the server
with a signed
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
terminology). Users may log into the authorization server directly
(e.g., via HTML forms), and session credentials there are stored in
cookies.

Client applications also can authenticate as a user to a Sourcegraph
server using OAuth2. After the client receives the authorization code
for the user from the authorization server, it authenticates using its
client credentials to obtain the user's access token.

An access token encodes the user's UID and the client ID that the
access token was generated for.


# ACLs and permissions

Currently, only simple ACLs are defined. Certain API methods, such as
those necessary to log in, are always accessible to everyone. All
other API methods may be restricted to only logged-in users.

