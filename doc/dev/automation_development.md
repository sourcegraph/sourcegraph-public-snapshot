# Developing automation

Automation features require creating changesets (PRs) on code hosts. If you are not part of the Sourcegraph organization, we recommend you create dummy projects to safely test changes on so you do not spam real repositories with your tests. If you _are_ part of the Sourcegraph organization, we have an account set up for this purpose.

## GitHub account safe for testing changeset creation

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

## Starting up your environment 

1. run `./enterprise/dev/start.sh` — Wait until all ~187 repositories are cloned.
2. Follow the [user guide on creating campaigns](../user/automation.md). Careful: if you use something like `repo:github` as a `scopeQuery` that will match all your repos. It takes a while to preview/create a campaign but also helps a lot with finding bugs/errors, etc.
