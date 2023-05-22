## Cody Slack (experimental)

Cody-Slack is an experimental Slack bot designed to bring the power of Cody to your Slack workspace. The main goal is to provide a useful developer experience tool that shares knowledge directly within Slack, interacts with different data stores (Sourcegraph repo, handbook, docs website), uses tools like GitHub CLI to do things and ultimately boosts productivity.

## Architecture

Cody-Slack's architecture is similar to the VSCode extension, as it's built on top of the `cody-shared` package. The Slack bot responds only to `app_mention` messages and uses all messages from a Slack thread where it's mentioned as a prompt context to improve responses.

Currently, key settings like repo (`sourcegraph/sourcegraph`), embeddings type (`blended`), and serverEndpoint (`s2`) are hardcoded. The Slack bot leverages Sourcegraph embeddings endpoint and Slack thread messages to create prompts. It then streams results back to the Slack thread, throttling message updates to avoid bumping into the Slack rate limiter.

By default, the Slack bot uses the Claude model, but it also has an implemented OpenAI completion client for future testing.

## Local Development

Use the `sg run cody-slack` command and provide all required environment variables via `sg.config.overwrite.yaml`. You can use `ngrok` to expose the local port to the internet and set the URL as [the Events Request URL](https://api.slack.com/apis/connections/events-api#request-urls) in the Slack configuration. To avoid affecting the production version of the Slack bot, create your own application for local development purposes.

An alternative option is to use [Socket Mode](https://api.slack.com/apis/connections/socket), which automatically connects to the Slack backend. However, Socket Mode has proven unreliable, as it regularly loses events, making local development challenging. Several GitHub issues are related to this problem.

Is there a better way to approach the local development of the Slack bot with the Events API? File a PR to update this section if you know about it!

## Deployment

The production deployment is rather simple: all code is bundled together using esbuild into a single JavaScript file (`pnpm build`), which is then uploaded to a Heroku eco dyno (`pnpm release`). The second command works only for @valerybugakov locally since the deployment is configured for his account. While this approach is suitable for prototyping with a 5-second re-deployment time, it would be great to host the Slack bot on Sourcegraph infrastructure (help wanted 👋).
