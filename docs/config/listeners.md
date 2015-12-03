+++
title = "Listeners and endpoints"
description = "Configure your Sourcegraph server port bindings"
+++

Sourcegraph exposes 3 services over HTTP, all multiplexed on the
same port:

* HTTP for the Web app
* HTTP for the REST API
* HTTP/2 for the gRPC API

You can [configure Sourcegraph to use TLS]({{< relref
"config/https.md" >}}) to make these services available over HTTPS,
in addition to or instead of HTTP.

Note: The examples in this document use `http://` URLs, because that is the
default. On the public Internet, you are strongly recommended to
[enable TLS]({{< relref "config/https.md" >}}) and use `https://` URLs
to avoid leaking credentials on the wire.

# Listen ports

You can configure the HTTP listener with:

```
[serve]
HTTPAddr = :3080
```

The port can have the following formats:

* `:port` to listen on all interfaces (`:3080`, for example)
* `addr:port` to listen on a single address (`10.1.2.3:3080`, for example)

## Using privileged ports 80 and 443 (Linux)

It's easiest when the Sourcegraph server is accessible on HTTP port 80
or HTTPS port 443, so that Web clients don't have to specify an
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
```

Or, if [using TLS]({{< relref "config/https.md" >}}):

```
[serve]
HTTPSAddr = :443
CertFile = my.crt
KeyFile = my.key
```

Note: This is not supported on all Linux systems and may introduce
additional security considerations (use at your own risk).
