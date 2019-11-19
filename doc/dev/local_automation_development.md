# Testing automation locally

1. Find the GitHub sd9 user in 1Password
2. Copy the Automation Testing Token
3. Change your `dev-private/enterprise/dev/external-services-config.json` to only contain a GitHub external service config with the token, like this:

```json
{
  "GITHUB": [
    {
      "authorization": {},
      "url": "https://github.com",
      "token": "<TOKEN>",
      "repositoryQuery": ["affiliated"]
    }
  ]
}
```
4. run `NO_KEYCLOAK=1 ./enterprise/dev/start.sh` — Wait until all ~187 repositories are cloned.
5. Follow the [user guide on creating campaigns](/user/automation). Careful: if you use something like `repo:github` as a `scopeQuery` that will match all your repos. It takes a while to preview/create a campaign but also helps a lot with finding bugs/errors, etc.
