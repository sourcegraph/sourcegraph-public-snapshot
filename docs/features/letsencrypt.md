# Using Let's Encrypt with Sourcegraph

The following instructions describe how to use [Let's Encrypt](https://letsencrypt.org) with Sourcegraph.

## Generate Certificate

1. These commands _must_ be run on the server you'll be running Sourcegraph on.
1. Ensure nothing is already running on port 80 (e.g. `sudo stop src` or `docker stop src` first).
1. `git clone https://github.com/letsencrypt/letsencrypt`
1. `cd letsencrypt`
1. `./letsencrypt-auto certonly --standalone --agree-tos --email admin@mysite.com -d src.mysite.com`
  - Note: Let's Encrypt certificates expire after 90 days, it is advised that you re-run the above commands every 60 days.

## Configuring Sourcegraph

- Follow the instructions for [configuring HTTPS and TLS](https://src.sourcegraph.com/sourcegraph/.docs/config/https/).
- For example:

```
[serve]
; Points to a TLS certificate and key.
CertFile = /etc/letsencrypt/live/mysite.com/fullchain.pem
KeyFile = /etc/letsencrypt/live/mysite.com/privkey.pem

; Be sure that your AppURL is "https://...".
AppURL = https://src.mysite.com

; Redirect "http://..." requests to "https://...".
RedirectToHTTPS = true
```

## Installing Root CA certificate

You will need to install the proper Root CA certificate onto your server in order
for the Sourcegraph frontend to connect to itself via SSL. In the case of
Let's Encrypt, this is the IdenTrust CA as [they cross-sign all Let's Encrypt certs](https://letsencrypt.org/certificates/).

1. (Docker-only): use `docker exec -it src bash` to get a shell.
1. `vi /usr/local/share/ca-certificates/identrust.crt`
1. Paste the [IdenTrust DST Root CA X3](https://www.identrust.com/certificates/trustid/root-download-x3.html) contents.
1. Type `:wq` + press enter to save the file.
1. `update-ca-certificates`
1. Restart Sourcegraph via `service src restart` or `docker src restart`.


### Docker

If you're running the Sourcegraph inside a Docker container, you'll need to copy
the newly generated PEM files into the container.

1. The files in `/etc/letsencrypt/live/mysite.com` are symlinks so find out
which files you should copy: `ls -l /etc/letsencrypt/live/mysite.com`.
1. Copy your certificates into the container:
  - `docker cp /etc/letsencrypt/archive/mysite.com/fullchain3.pem src:/etc/sourcegraph/fullchain.pem`
  - `docker cp /etc/letsencrypt/archive/mysite.com/privkey3.pem src:/etc/sourcegraph/privkey.pem`
1. Grab an interactive shell:
  - `docker exec -it src bash`
1. `exit && docker src restart`

## Troubleshooting

### Error "Failed to connect to host for DVSNI challenge"

Let's Encrypt tries to verify your ownership of the server via ports :80 and :443,
so confirm that your DNS settings correctly point to your server and that you
are not being routed through e.g. CloudFlare or other HTTPS middleware
providers while running the above commands.

## Known Issues

In the future, Sourcegraph will seamlessly integrate with Let's Encrypt in order
to automatically renew certificates for you. Until then, you'll need to rerun
the above `./letsencrypt-auto certonly` step to renew your certificate at least
every 90 days to avoid your HTTPS service being interrupted.

In the case of Docker, you'll also need to copy the new certificates into the
container before restarting the Sourcegraph server.
