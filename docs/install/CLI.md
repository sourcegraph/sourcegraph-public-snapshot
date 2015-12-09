+++
title = "Command line interface"
linktitle = "CLI"
+++

The Sourcegraph CLI (`src`) is bundled with binary which runs your Sourcegraph server.

To install it, [follow the local installation instructions]({{< relref "getting-started/local.md" >}}).

# Usage

You may use the CLI to search code, create and merge changesets, and more.

First, connect to a running Sourcegraph server:

```
src --endpoint=http://localhost:3080 login
```

Then follow the help prompt to find a command:

```
src -h
```

{{< ads_conversion >}}
