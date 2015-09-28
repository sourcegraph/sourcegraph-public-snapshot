+++
title = "Listeners and endpoints"
+++

Sourcegraph has three main services, which all listen on all
interfaces by default:

* Port 3000: HTTP for the Web app
* Port 3001: HTTP/2 (and HTTPS, if [TLS is enabled]({{< relref "config/https.md" >}})) for the Web app
* Port 3100: HTTP/2 (and HTTPS, if [TLS is enabled]({{< relref "config/https.md" >}})) for the [gRPC](http://www.grpc.io) API

The examples in this document use `http://` URLs, because that is the
default. On the public Internet, you are strongly recommended to
[enable TLS]({{< relref "config/https.md" >}}) and use `https://` URLs
to avoid leaking credentials on the wire.

# Listen ports

You can configure the listeners using the settings shown below.

```
[serve]
HTTPAddr = :3000
Addr = :3001
GRPCAddr = :3100
```

The values can have the following formats:

* `:port` to listen on all interfaces (`:3000`, for example)
* `addr:port` to listen on a single address (`10.1.2.3:3000`, for example)

Make sure that the app URL and endpoint URLs (explained below) refer
to the externally accessible URLs of the listeners defined here. For
example, if your `HTTPAddr` is `:7000`, and you expect Web clients to
access the server directly via HTTP (and not via a proxy or using port
forwarding), then the app URL probably should be
`http://example.com:7000`. If you are using port forwarding, a load
balancer, or an HTTP reverse proxy, then adjust accordingly.

## Using ports 80 and 443 (Linux)

It's easiest when the Sourcegraph server is accessible on HTTP port 80
and HTTPS port 443, so that Web clients don't have to specify an
alternate port. But Sourcegraph shouldn't run as the root user. You
could set up iptables port forwarding to achieve this, but a simpler
way to allow Sourcegraph to listen directly on ports 80 and 443 is to
grant it the `cap_net_bind_service` Linux capability:

```
sudo setcap cap_net_bind_service=+ep /usr/bin/src
```

Then you can configure the listeners as follows:

```
[serve]
HTTPAddr = :80
Addr = :443
```

Note: This is not supported on all Linux systems and may introduce
additional security considerations (use at your own risk).


# App URL

Both the Sourcegraph server itself and external API clients (such as
the `src` CLI) need to know how to reach the server by its publicly
available URL.

## Server

On the server, this is configured as follows:

```
[serve]
# Sets the URL at which external clients can access this Sourcegraph server.
AppURL = http://example.com
```

## `src` CLI

In the `src` CLI, to communicate with a specific Sourcegraph server, use:

```
src --endpoint http://example.com <command...>
```

Usually you need to authenticate with the server before you can
perform API operations. Log in by running:

```
src --endpoint http://example.com login
```

After authenticating, future CLI operations will be performed against
the endpoint you specified. Your authentication information is saved
in `~/.src-auth`.

# API endpoints

The server must also know how to contact its own API, and what
externally accessible API endpoint URLs to publish. Often it can
determine these from the app URL and listen ports, but sometimes you
need to specify them manually (if you are using port forwarding, for
example).

```
[serve]
; Set these to the externally accessible endpoints.
; Note the trailing slash in the HTTPEndpoint!
HTTPEndpoint = http://example.com/api/
GRPCEndpoint = http://example.com:3100

[Client API endpoint]
; Set this to the internally accessible gRPC endpoint (which is usually,
; but not always, the same as the externally accessible gRPC endpoint).
GRPCEndpoint = http://example.com:3100
```
