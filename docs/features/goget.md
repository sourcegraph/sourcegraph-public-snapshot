+++
title = "Go get support"
+++

All repositories hosted on Sourcegraph automatically support [go get](https://golang.org/cmd/go/#hdr-Remote_import_paths). Only basic configuration of the server is required, specifically you must:

1. [Configure AppURL and DNS]({{< relref "config/appurl-dns.md" >}}); `go get` requires a domain name to work.
1. [Configure HTTPS and TLS]({{< relref "config/https.md" >}}); otherwise `go get --insecure` must be used.

## Public Repositories

First [configure Sourcegraph to be public]({{< relref "config/public.md" >}}) if you haven't already, and then simply `go get src.example.com/my/pkg`.

## Private Repositories

At this time, `go get` does not work with private Sourcegraph repositories. However, you can set it up manually:

- `git clone https://src.example.com/my/repo $GOPATH/src/src.example.com/my/repo`

After which all normal `go get` operations will work (e.g. `go get -u src.example.com/my/repo/...` to get latest updates).
