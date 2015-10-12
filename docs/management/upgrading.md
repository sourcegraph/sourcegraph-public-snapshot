+++
title = "Upgrading Sourcegraph"
navtitle = "Upgrading"
+++

When an updated version of Sourcegraph is available for your server, a
notice will appear in the footer.

# Upgrading

1. Log into the server running Sourcegraph and run `sudo src
   selfupdate` to update the binary to the latest version.
1. If you used the cloud provider installation instructions, re-grant
   the capability to let Sourcegraph listen on privileged ports (if
   needed): `sudo setcap cap_net_bind_service=+ep /usr/bin/src`
1. If running Sourcegraph as a system service: `sudo restart src`
