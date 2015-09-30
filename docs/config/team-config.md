+++
title = "Team configuration"
+++

# Update hostname / DNS

Sourcegraph will run with configuration found at `/etc/sourcegraph/config.ini`.

By default, your Sourcegraph server will be accessible by the public IP address of its
host machine, e.g. `http://<ip-address>`

To change the URL of your Sourcegraph instance (e.g. `http://src.mycompany.com`)
you may update the configuration file:

```
; Change this from http://<ip-address>
AppURL = http://src.mycompany.com
```

# Access control

Follow instructions to [restrict access to teammates]({{< relref "config/access-control.md" >}})
or [make your Sourcegraph publicly accessible]({{< relref "config/public.md" >}}).

# TLS

[Enabling TLS]({{< relref "config/https.md" >}}) is *strongly recommended*,
to avoid leaking credentials.
