+++
title = "GitHub.com"
+++

## Public repositories

To mirror public GitHub.com repositories on a Sourcegraph instance, run this command:

```
src repo create -m --clone-url https://github.com/mycompany/project <repo-name>
```

## Private repositories

Sourcegraph can import private GitHub.com code, enabling a limited set of
Sourcegraph features for the externally hosted repository.

To use this feature, you must have a [PostgreSQL backend]({{< relref "config/localstore.md" >}})
for your Sourcegraph server. To enable this feature, follow these steps:

- [Create a GitHub developer application](https://github.com/settings/developers)
- Set the "Authorization callback url" to `http://$APP_URL/github-oauth/receive`, where
  `$APP_URL` is the URL at which you access your Sourcegraph server.
- Get the client id and client secret. Set these env vars in your Sourcegraph server's
  environment:
  
  ```
  export GITHUB_CLIENT_ID=[client_id]
  export GITHUB_CLIENT_SECRET=[client_secret]
  ```

- Make sure you are running the server with a PostgreSQL database backend. Also, you may need to
  run the db migration steps listed in `dbutil2/MIGRATE.md` if you have updated your server
  from an older version (<= 0.13.6)
- Append the flag `--private-mirrors` to your server config.
- Restart your server.

With this feature, you will be able to link your Sourcegraph user account with
your GitHub account via OAuth2, after which you can mirror your private GitHub repos on
your Sourcegraph server. The repo permissions are synced with GitHub, so everyone
who wants to access the private repo on Sourcegraph will have to link their account
with GitHub.

### Migrating from personal access tokens

If you used an earlier version of Sourcegraph to mirror private GitHub repos using personal
access tokens, you will not be able to continue using that feature. Please follow the steps
mentioned above to set up the private mirrors feature for your team.
