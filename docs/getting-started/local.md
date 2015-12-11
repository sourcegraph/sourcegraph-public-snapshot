+++
title = "Installing Sourcegraph locally"
linktitle = "on your desktop/laptop"
+++

Sourcegraph in an all-in-one binary (`src`) which includes:

- a Sourcegraph server with a Web app and Git server
- a command line interface

*You must separately install **Code Intelligence** to enable Sourcegraph's advanced features.*

# Install Sourcegraph

{{< local_install >}}

# Add Code Intelligence

{{< toolchain_install >}}

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
