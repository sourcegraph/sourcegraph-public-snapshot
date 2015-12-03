+++
title = "AppURL and DNS"
description = "Update your Sourcegraph server URL"
+++

If you installed Sourcegraph using one of the standard distribution or cloud provider packages,
Sourcegraph will run with configuration found at `/etc/sourcegraph/config.ini`.

By default, your Sourcegraph server will be accessible at the public IP address of its
host machine, e.g. `http://<ip-address>`

To change the URL of your Sourcegraph server (e.g. `http://src.mycompany.com`),
set your DNS records to point to your server's IP address and update the configuration file:

```
; Change this from http://<ip-address>
AppURL = http://src.mycompany.com
```

After updating your config, restart the `src` process:

```
sudo restart src
```

If you have already logged in and authorized your Sourcegraph server with OAuth,
you'll also need to reset the redirect URI of your instance. First, get your instance's
`IDKey`:

```
src --endpoint http://src.mycompany.com meta config
```

Then, update the server redirect URI:

```
src --endpoint https://sourcegraph.com login
src registered-clients update --redirect-uri http://src.mycompany.com <IDKey>
# to reset src's default --endpoint, run "src --endpoint http://src.mycompany.com login"
```

***IMPORTANT***: To re-access your Sourcegraph server after running the step above, you'll
 need to clear your browser cookies. First make sure you've logged out of and closed any browser
 window to your server.

 Finally, go to `http://src.mycompany.com` to access your newly named Sourcegraph instance.

# Access control

Follow instructions to [restrict access to teammates]({{< relref "management/access-control.md" >}})
or [make your Sourcegraph publicly accessible]({{< relref "config/public.md" >}}).

# TLS

[Enabling TLS]({{< relref "config/https.md" >}}) is *strongly recommended* to avoid sending
credentials in plaintext.

# Inviting team members

In conclusion, to invite a team member (after following the above configuration steps) you just need to grant them access via <code>src access grant USERNAME</code> and give them a link to your Sourcegraph instance (the <code>AppURL</code> you configured).
