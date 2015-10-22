+++
title = "Security"
+++

NOTE: The code does not yet implement all of the requirements and
models described in this document.


# Principles

Sourcegraph's core security principles are:

* **Small attack surface.** The attack surface must be as small as
  possible. Only the gRPC service (`server/...` packages) has direct
  access to sensitive resources (e.g., the filesystem with
  repositories, PostgreSQL, the build queue files, etc.). The app,
  HTTP API, and worker are all mere API clients that may only access
  the server's resources through the public gRPC API. If the gRPC
  service denies permission to perform an action, these API clients
  have no way to circumvent that. As described in
  [OAuth2.md]({{< relref "dev/OAuth2.md" >}}), these clients are authenticated with the
  server using a shared secret, but the important part is that they
  have no out-of-band access to the server's resources.

  Suppose instead that the app, HTTP API, and worker had direct access
  to the server's resources. In some ways this would be simpler: if
  the worker had access to the filesystem where the repositories
  lived, it could directly clone repositories during builds without
  worrying about authentication. However, this means we would need to
  consider the worker's entire API as an attack surface for the larger
  system. By constraining the worker's access to only that allowed by
  the gRPC API methods, the security of the larger system is easier to
  reason about.

* **Security by default.** This is important because the Sourcegraph
  server is assumed to be publicly available on the Internet.

* **Be robust to mistakes.** Design APIs and write code such that
  bypassing authorization checks is very hard or impossible.

  For example, consider the following two APIs. API 1 has a method
  `SendMoney(amount int, account string, authorized bool)` and relies
  on the caller to determine whether the operation is authorized. API
  2 has a method `SendMoney(amount int, account string, authorization
  Token)` and checks that the authorization token permits the
  operation itself. It's far harder for the caller to bypass API 2's
  authorization check because it would require explicit effort to
  construct a valid authorization token for the operation, which would
  raise red flags during code review.


# Assumptions

Assume that all Sourcegraph servers--gRPC, Web app, HTTP API, and
git--are publicly accessible on the Internet. Even though many users
will run Sourcegraph on internal, non-public servers, we must not rely
on security provided by network configuration because that is out of
our control.

Assume that repositories explicitly added to a Sourcegraph instance
are non-malicious. Building them may execute code inside the
repositories (such as npm package.json scripts, Maven pom.xml tasks,
etc.). This is consistent with how many package managers work. Also,
it's likely that the same user who added a repository to their
Sourcegraph would also depend on libraries from that repository, in
which case they would need to trust the code.

Assume that every attacker has full access to the Sourcegraph source
code.


# Informal specification

A Sourcegraph server consists of the following externally accessible
systems:

* the gRPC API server, which exposes the public API
* the Web app
* the HTTP API
* the build worker
* the git repository host (which supports `git push` and `git pull`)

Because all access to sensitive resources must go through the gRPC API
(no other systems have out-of-band access to the resources), we can
focus on the gRPC API's security in isolation and then consider the
other systems.


## gRPC API

WORK-IN-PROGRESS.

See also server/doc.go (WIP).

### Actor

Every gRPC call is originated by some agent, called an **actor**. An
actor can be:

* completely unauthenticated,
* client-authenticated but not as any particular user, or
* both client- and user-authenticated.

Here, "client" refers to an OAuth2 client (see
[OAuth2.md]({{< relref "dev/OAuth2.md" >}}) for more info).

Depending on the level of authentication, the following information is
available about the current actor:

* Client ID of authenticated OAuth2 client
* UID of authenticated user
* Domain (hostname) of authenticated user
* List of authorized scopes

Server code (in `server/...`) may access the actor using
`auth.ActorFromContext(ctx)`. The actor is not available from the
other systems (Web app, HTTP API, worker, and git repository host)
because it is an implementation detail of the server. The other
systems may pass client and user tokens with all gRPC calls they
invoke, but they should treat those tokens as opaque. If they need to
identify the current client or user, they may call the gRPC methods
`Auth.Identify` or `RegisteredClients.GetCurrent`, respectively, using
their tokens.

## Web app

TODO

## HTTP API

TODO

## Worker

TODO

## Git repository host

TODO
