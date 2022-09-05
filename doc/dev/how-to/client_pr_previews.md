# Exploring client changes with PR previews

We have app previews for each PR with **only client changes**. It means that everyone can easily explore UI updates and play with new client features by clicking a link in the PR description instead of pulling a branch and running the application locally!

## Under the hood

A link to a deployed preview is automatically added to the bottom of the PR description. [Under the hood](https://sourcegraph.com/search?q=context:global+repo:sourcegraph/sourcegraph%24+Get+service+id+of+preview+app+on+render.com&patternType=literal), it uses a production build of the web application running on the standalone web server:

```sh
dev/ci/yarn-build.sh client/web
yarn workspace @sourcegraph/web serve:prod
```

The standalone web server is deployed as [a web service](https://render.com/docs/web-services) to [render.com](https://render.com/) via [a Buildkite step](https://sourcegraph.com/search?q=context:global+repo:sourcegraph/sourcegraph%24+prPreview%28%29&patternType=literal).

## How to login

[The dogfood instance](https://k8s.sgdev.org/) API is used as the backend, so use your k8s username and password credentials to log in to the preview version of the application. Use the dogfood instance to sign up if you do not have one.

Other resources:

- [Testing in Dogfood](./testing_in_dogfood.md)

## Why does it work for client-only pull requests?

We do not create client PR previews [if Go or GraphQL is changed](https://sourcegraph.com/search?q=context:global+repo:sourcegraph/sourcegraph%24+%21c.Diff.Has%28changed.Go%29&patternType=literal) in a PR to avoid confusing preview behavior because only Client code is used to deploy application preview. Otherwise, the PR preview would sometimes show incomplete features with the updated client code, which is not yet supported by the backend.

## Why does a search query fail with an error?

The preview app is deployed with `SOURCEGRAPHDOTCOM_MODE=false`, which means that the user should be authenticated to use all web application features similar to [the dogfood instance](https://k8s.sgdev.org/). Make sure that you're logged in. If it doesn't fix the issue, please report it in Slack.

## Why is my preview inactive?

Previews are made inactive, when they exceeds the preview liftime. This is done to save resource (which is required to keep preview active over an extended period) from beign spent on redundant previews.

## Where to find the script that cleans up previews?

The preview cleanup script is located [here](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/ci/render-pr-preview-cleanup.sh)

## What is the default cleanup schedule and the default preview lifetime?

The Default cleanup schedule runs every 12th hour, where default preview lifetime is 5 days. This can be modified through the `-e` (e.g `-e 5` for 5 days) flag passed to the preview script. Preview would also be removed once a PR gets closed or merged.
