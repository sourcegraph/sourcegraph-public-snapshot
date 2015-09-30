+++
title = "AppURL and DNS"
+++

# Update hostname

If you installed using one of the standard distribution or cloud provider packages,
Sourcegraph will run with configuration found at `/etc/sourcegraph/config.ini`.

By default, your Sourcegraph server will be accessible at the public IP address of its
host machine, e.g. `http://<ip-address>`

To change the URL of your Sourcegraph server (e.g. `http://src.mycompany.com`),
update the configuration file:

```
; Change this from http://<ip-address>
AppURL = http://src.mycompany.com
```

Then after updating your config, restart the `src` process:

```
sudo restart src
```

If you have already authorized your Sourcegraph server with OAuth, you'll also
need to reset the redirect URI of your instance:

```
src registered-clients update --redirect-uri http://src.mycompany.com SOURCEGRAPH-CLIENT-ID
```

# Access control

Follow instructions to [restrict access to teammates]({{< relref "config/access-control.md" >}})
or [make your Sourcegraph publicly accessible]({{< relref "config/public.md" >}}).

# TLS

[Enabling TLS]({{< relref "config/https.md" >}}) is *strongly recommended*,
to avoid leaking credentials.
