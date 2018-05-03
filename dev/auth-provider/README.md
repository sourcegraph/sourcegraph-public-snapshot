# Dev auth provider

[Keycloak](https://www.keycloak.org) is an authentication provider that we use in development to test Sourcegraph's support for OpenID and SAML. It serves the same role as OpenID and SAML providers on Okta, OneLogin, G Suite, etc., but it runs locally and is easier to autoconfigure for use with your local dev server.

## Using Keycloak in local dev

1.  Run Keycloak: `./keycloak.sh`
1.  To make your local Sourcegraph dev server use OpenID or SAML provided by Keycloak, uncomment the relevant section in `dev/config.json` or on the site admin configuration page.

> If Keycloak is useful, we will add it to the dev/Procfile so it always runs in local dev.

## Configuring Keycloak, adding users, etc.

If you need to edit client or user information and want to persist your changes:

1.  Start Keycloak, if you haven't already. See the above section for steps. The `keycloak` Docker container should be running.
1.  Edit the JSON files in `config/` as needed.
1.  Run `RESET=1 scripts/configure-keycloak.sh` to clobber the existing configuration with the `config/*.json` files' configuration.

Not sure how to edit the JSON to achieve your desired outcome? Use the Keycloak admin interface at http://localhost:3220/auth (login as `root`/`q`) to change configuration, and then export to JSON.
