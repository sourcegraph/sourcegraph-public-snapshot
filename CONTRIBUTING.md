# Contributing

## What contributions are accepted?

In short, we are open to nearly all contributions! We love feedback in all forms, issues, comments, PRs, etc!

Unless you feel confident your change will be accepted (trivial bug fixes, code cleanup, etc) you should first create an issue or a [Sourcegraph RFC](https://about.sourcegraph.com/handbook/communication/rfcs#external-contributors) (preferred for bigger changes) to discuss your change with us. This lets us all discuss the design and proposed implementation of your change, which helps ensure your time is well spent and that your contribution will be accepted.

> Exception: If you contribute functionality that already exists as a [paid Sourcegraph feature](https://about.sourcegraph.com/pricing/), we are unlikely to accept it. Consult us beforehand for a definitive answer. (We'll add more details about the process here, and they'll be similar to [GitLab's stewardship principles](https://about.gitlab.com/stewardship/#contributing-an-existing-ee-feature-to-ce).)

## Code of Conduct

All interactions with the Sourcegraph open source project are governed by the
[Sourcegraph Code of Conduct](https://handbook.sourcegraph.com/community/code_of_conduct).

## How to contribute

1. Select one of the issues labeled as [good first issue](https://github.com/orgs/sourcegraph/projects/210).
2. Clone the repo: `git clone https://github.com/sourcegraph/sourcegraph/`.
3. [Setup your development environment](https://docs.sourcegraph.com/dev/contributing) to run the project locally.
4. Before creating a Pull Request ensure that [recommended checks](https://docs.sourcegraph.com/dev/contributing) pass locally. We're actively working on making our CI pipeline public to automate this step.
5. **IMPORTANT:** Once you have a pull request ready to review, the 'verification/cla-signed' check will be flagged, and you will be prompted to sign the CLA with a link provided by our bot. Once you sign, add a comment tagging `@natectang`. After that your pull request will be ready for review.
6. If one or more people ask for the same issue, it will be assigned to the first person who asked.
7. If you have any questions, please [refer to the docs first](https://docs.sourcegraph.com/). If you don’t find any relevant information, mention the issue author.
8. Issue author will try to provide guidance. Sourcegraph always works in async mode. We will try to answer as soon as possible, but please keep time zones differences in mind.

## Relevant development docs

### Getting applications up and running

- [Getting Started Guide](https://docs.sourcegraph.com/dev/getting-started)
- [Troubleshooting section](https://docs.sourcegraph.com/dev/how-to/troubleshooting_local_development)

### How to write tests

- [How to write tests](https://docs.sourcegraph.com/dev/how-to/testing)
- [Testing principles](https://docs.sourcegraph.com/dev/background-information/testing_principles)
- [Testing web code](https://docs.sourcegraph.com/dev/background-information/testing_web_code)

### Pull Requests

- [How to structure](https://docs.sourcegraph.com/dev/background-information/code_reviews#what-makes-an-effective-pull-request-pr)
- [Size guidelines](https://about.sourcegraph.com/handbook/engineering/developer-insights#prefer-small-prs-lines)
- Git branch name convention: `[developer-initials]/short-feature-description`
- [Examples on Github](https://github.com/sourcegraph/sourcegraph/pulls?q=is%3Apr+label%3Ateam%2Ffrontend-platform)
