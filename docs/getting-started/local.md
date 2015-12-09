+++
title = "Installing Sourcegraph locally"
linktitle = "Local install"
+++

Sourcegraph in an all-in-one binary (`src`) which includes:

- a Sourcegraph server with a Web app and Git server
- a command line interface

# Instructions

{{< local_install >}}

## Add language support

Out-of-the box, Sourcegraph does not include Code Intelligence.
[Follow these instructions]({{relref "config/toolchains.md" >}}) to enable
language support on your Sourcegraph server.

## Add a git repository

Run these commands from to push a repository to Sourcegraph after you've created an admin
account on your Sourcegraph server:

```
src --endpoint=http://localhost:3080 login
src repo create project
git push http://localhost:3080/project master
```

# Next steps

To use Sourcegraph with your teammates, we recommend following our
[cloud installation instructions]({{< relref "getting-started/cloud.md" >}}).

Or, you can invite your teammates to join your local server!

1. [Find your computer's local IP address]
(http://stackoverflow.com/questions/13322485/how-to-i-get-the-primary-ip-address-of-the-local-machine-on-linux-and-os-x).
2. Set the `--app.url` flag when running your server (replacing `$LOCAL_IP_ADDRESS` with the result of step 1):
	```
	src serve --app.url=$LOCAL_IP_ADDRESS
	```
3. [Invite your teammates to your server]({{< relref "management/access-control.md" >}}).


**Warning:** your local IP address may change and your teammates will no longer be able to
access your server. [Deploy Sourcegraph to the cloud]({{< relref "getting-started/cloud.md" >}})
to easily set a static IP address & DNS records for your server.

{{< ads_conversion >}}
