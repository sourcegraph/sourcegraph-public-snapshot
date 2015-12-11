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

# Invite teammates

{{< cloud_cta >}} will deploy Sourcegraph on **AWS**, **DigitalOcean**, or **Google Cloud Platform** in minutes.

You can also invite teammates to a local Sourcegraph server if you're on the same network!

#### 1. [Find your computer's local IP address](http://stackoverflow.com/questions/13322485/how-to-i-get-the-primary-ip-address-of-the-local-machine-on-linux-and-os-x).
#### 2. Run `src serve --app.url=$LOCAL_IP_ADDRESS`
#### 3. [Invite your teammates]({{< relref "management/access-control.md" >}}).

**Warning:** your local IP address may change and your teammates will no longer be able to
access your server.

{{< ads_conversion >}}
