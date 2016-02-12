+++
title = "GitHub.com"
+++

## Public repositories

To mirror public GitHub.com repositories on a Sourcegraph instance, run this command:

```
src repo create -m --clone-url https://github.com/mycompany/project <repo-name>
```

## Private repositories

Sourcegraph can import private GitHub.com repositories, enabling a limited set of
Sourcegraph features for the externally hosted repository.

To use this feature, you must have a PostgreSQL backend for your Sourcegraph server.
To enable this feature, follow these steps:

- Create a GitHub developer application by visiting https://github.com/settings/developers
  and clicking on "Register new application".
- Set the "Authorization callback url" to https://APP_URL/github-oauth/receive, where
  APP_URL is the URL at which you access your Sourcegraph server.
- Get the client id and client secret. Set these env vars in your Sourcegraph server's
  environment:
  GITHUB_CLIENT_ID=[client_id]
  GITHUB_CLIENT_SECRET=[client_secret]
  (If you have a cloud install of Sourcegraph, you can find the env config file at
  `/etc/sourcegraph/config.env`)
- Make sure you are running the server with a PostgreSQL database backend. Also, you may need to
  run the db migration steps listed in `dbutil2/MIGRATE.md` if you have updated your server
  from an older version (<= 0.13.6)
- Append the flag `--auth.mirrors-next` to your server config. (If you have a cloud install of
  Sourcegraph, you can find the config flags file at `/etc/sourcegraph/config.ini`)
- Restart your server with `sudo restart src` on Ubuntu, or run with `src serve` on OS X.

With this feature turned on, you will be able to link your Sourcegraph user account with
your GitHub account via OAuth2, after which you can mirror your private GitHub repos on
your Sourcegraph server. The repo permissions are synced with GitHub, so everyone
who wants to access the private repo on Sourcegraph will have to link their account
with GitHub.

### Migrating from personal access tokens

If you used an earlier version of Sourcegraph to mirror private GitHub repos using personal
access tokens, you will not be able to continue using that feature and must switch to the
MirrorsNext feature outlined above. Please follow the steps mentioned above to set up this
feature for your team. All your private repo data will remain in Sourcegraph through the
transition; however, every user on your team will have to link their GitHub accounts in
order to access the private repositories via Sourcegraph.
