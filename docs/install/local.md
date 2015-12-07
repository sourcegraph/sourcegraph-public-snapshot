+++
title = "Local installation"
linktitle = "Local installation"
+++

Sourcegraph in an all-in-one binary (`src`) which includes:

- a Sourcegraph Core server
- a web interface
- a command line interface

# Install on Mac OS X

Sourcegraph is supported on OS X 10.8+. Install using one of the following methods:

* `curl -sSL https://sourcegraph.com/.download/install.sh | bash`
* [Download a gzipped static binary for Mac OS X](https://sourcegraph.com/.download/latest/darwin-amd64/src.gz)
* (Homebrew coming soon)

Next, run `src serve` and visit [http://localhost:3080](http://localhost:3080).

# Install on Linux

Linux installations will create a configuration file at `/etc/sourcegraph/config.ini` and
an upstart script which runs your Sourcegraph server.

## Ubuntu 12.04/14.04

Install using one of the following methods:

* `wget -O - https://sourcegraph.com/.download/install.sh | bash`
* Download [src.deb (64-bit)](https://sourcegraph.com/.download/latest/linux-amd64/src.deb)
and install with `sudo dpkg -i src.deb`

When done, run `sudo restart src` and visit [http://localhost:3080](http://localhost:3080)
(or an alternate address, if you modified the configuration).

## CentOS 6.3 or Red Hat 7.1

Install using one of the following methods:

* `curl -sSL https://sourcegraph.com/.download/install.sh | bash`
* Download [src.rpm (64-bit)](https://sourcegraph.com/.download/latest/linux-amd64/src.rpm)
and install with `sudo yum -y src.rpm`

When done, run `src serve` and visit [http://localhost:3080](http://localhost:3080)
(or an alternate address, if you modified the configuration).

# Next steps

* [Install Code Intelligence]({{< relref "config/toolchains.md" >}}) on your local Sourcegraph server

or

* [Getting started with Sourcegraph for your team]({{< relref "getting-started/index.md" >}}) for scripts to deploy
Sourcegraph on the cloud with Code Intelligence enabled out of the box

{{< ads_conversion >}}
