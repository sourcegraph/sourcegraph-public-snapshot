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

We provide [automated cloud installation]({{< relref "getting-started/cloud.md" >}})
on AWS, DigitalOcean, and Google Cloud Platform so you can set up a team server in
minutes.

You can also invite your teammates to your local Sourcegraph server if you're all on the same network!

#### 1. [Find your computer's local IP address](http://stackoverflow.com/questions/13322485/how-to-i-get-the-primary-ip-address-of-the-local-machine-on-linux-and-os-x).
#### 2. Run `src serve --app.url=$LOCAL_IP_ADDRESS`
#### 3. [Invite your teammates]({{< relref "management/access-control.md" >}}).

**Warning:** your local IP address may change and your teammates will no longer be able to
access your server.

[Use automated cloud installation to deploy Sourcegraph for your team in minutes.]({{< relref "getting-started/cloud.md" >}})

{{< ads_conversion >}}
