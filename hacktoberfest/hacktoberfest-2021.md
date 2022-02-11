# Welcome Hacktoberfest participants

Thank you for checking out the [Sourcegraph](https://sourcegraph.com) repo, and for considering contributing to this project!

If you need guidance on contributing to open source, please review the tutorial [How To Create a Pull Request on GitHub](https://www.digitalocean.com/community/tutorials/how-to-create-a-pull-request-on-github) or watch the video [Your First Pull Request with Lyn Muldrow](https://www.youtube.com/watch?v=jZtECuvNRiw).

You can also join our [Sourcegraph Community Space](https://srcgr.ph/join-community-space) to ask questions and network.

For additional guidance, review the [official Hacktoberfest site](https://hacktoberfest.digitalocean.com/).

_Hacktoberfest contributions are limited to what is described below. All other pull requests will be marked as **invalid**._

## Step 1 — Choose an issue

Select one of the issues labeled as [good first issue](https://github.com/orgs/sourcegraph/projects/210).

Each of these issues is either a frontend code issue in the Sourcegraph OSS project or a troubeshooting technical tutorial for educational purposes. **If you want to solve a technical tutorial issue, please follow [this guide instead](https://github.com/sourcegraph/learn/blob/main/docs/hacktoberfest-2021.md)**.

Once you've chosen an issue, **comment on it to announce that you will be working on it**, making it visible for others that this issue is being tackled. If you end up not creating a pull request for this issue, please delete your comment.

### Can I pick up this issue?

All open issues are not yet solved. If the task is interesting to you, take it and feel free to do it. There is no need to ask for permission or get in line. Even if someone else can do the task faster than you, don't stop - your solution may be better. It is the beauty of Open Source!

## Step 2 - Fork the Sourcegraph repository

If you have the GitHub command line tool:

`gh repo fork https://github.com/sourcegraph/sourcegraph/ --clone=true`

Otherwise, [fork sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph/fork) then `git clone git@github.com:<your-username>/sourcegraph.git`.

For more information about forking a repository, please read the [GitHub docs](https://docs.github.com/en/get-started/quickstart/fork-a-repo#cloning-your-forked-repository).

## Step 3 - Set up your local development environment

Follow our [Frontend contribution guidelines](https://docs.sourcegraph.com/dev/contributing/frontend_contribution) to set up your local development environment so you can contribute to Sourcegraph's frontend codebase.

Check our [troubleshooting section](https://github.com/sourcegraph/sourcegraph/blob/main/doc/dev/index.md#troubleshooting) if you run into problems. If they persist, reach out to us in the #help channel of the [Sourcegraph Community Space](https://srcgr.ph/join-community-space) and we'll help you out.

(optional) We also highly recommend adding the [Sourcegraph browser extension](https://docs.sourcegraph.com/integration/browser_extension) to your browser. The open-source Sourcegraph browser extension adds code intelligence to files and diffs on GitHub and other code hosts and it creates a search shortcut in your URL location bar. The extension will make it easier for you to read and understand any code base you contribute to.

## Step 4 — Start coding on your solution

Create a branch for your contribution.

Write the code that will solve the issue and don't forget to add [tests](https://docs.sourcegraph.com/dev/how-to/testing).

## Step 5 — Open a pull request

Before creating a Pull Request ensure that [recommended checks](https://docs.sourcegraph.com/dev/contributing/frontend_contribution#ci-checks-to-run-locally) pass locally. We're actively working on making our CI pipeline public to automate this step.

When you are satisfied, open your pull request referencing the issue that you are resolving. **Use the word `hacktoberfest` in your pull request description** to increase its discoverability.

**IMPORTANT:** Once you have a pull request ready to review, the `verification/cla-signed` check will be flagged, and you will be prompted to sign the CLA with a link provided by our bot. Once you sign, add a comment tagging `@sourcegraph/hacktoberfest-reviewers`. After that, your pull request will be ready for review.

## Further info

If you have any questions, please [refer to the docs first](https://docs.sourcegraph.com/). If you don’t find any relevant information, mention the issue author. You can also ask questions in the [Slack Community Space #hacktoberfest channel](https://srcgr.ph/join-community-space).

The issue author will try to provide guidance. Sourcegraph always works in async mode. We will try to answer as soon as possible, but please keep time zones differences in mind.

Please note that maintainers will make the ultimate decision of whether or not pull requests are merged into the project.

## Rewards

We’ll send Sourcegraph swag to the first 50 people whose PRs get merged and you'll receive a cool Sourcegraph Hacktoberfest Contributor badge to show case on your GitHub, website or social media. In addition, the top contributor will have a pair-programming session with one of our engineers, streamed live on Twitch!

_Thank you for considering contributing to [Sourcegraph](https://sourcegraph.com)!_
