# Repository webhooks

# Webhook for manually telling Sourcegraph to update a repository

By default, Sourcegraph polls code hosts to keep repository contents up to date. It uses intelligent heuristics like average update frequency to determine the polling frequency per repository.

Polling, however, falls short in cases where immediate updates are desired or when the number of repositories causes significant load on the code host.

To address this, there is a repository update webhook that lets an external service (e.g., a code host) trigger a repository update on Sourcegraph. The webhook is authenticated using access tokens.

Here's an example using curl.

```bash
curl -XPOST -H 'Authorization: token $ACCESS_TOKEN' $SOURCEGRAPH_ORIGIN/.api/repos/$REPO_URI/-/refresh
```
