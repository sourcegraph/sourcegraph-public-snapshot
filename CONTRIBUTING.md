# Contributing

## What contributions are accepted?

In short, we are open to nearly all contributions! We love feedback in all forms, issues, comments, PRs, etc!

Unless you feel confident your change will be accepted (trivial bug fixes, code cleanup, etc) you should first create an issue or a [Sourcegraph RFC](https://about.sourcegraph.com/handbook/communication/rfcs#external-contributors) (preferred for bigger changes) to discuss your change with us. This lets us all discuss the design and proposed implementation of your change, which helps ensure your time is well spent and that your contribution will be accepted.

> Exception: If you contribute functionality that already exists as a [paid Sourcegraph feature](https://about.sourcegraph.com/pricing/), we are unlikely to accept it. Consult us beforehand for a definitive answer. (We'll add more details about the process here, and they'll be similar to [GitLab's stewardship principles](https://about.gitlab.com/stewardship/#contributing-an-existing-ee-feature-to-ce).)

## Code of Conduct

All interactions with the Sourcegraph open source project are governed by the
[Sourcegraph Code of Conduct](https://handbook.sourcegraph.com/community/code_of_conduct).

## How to contribute

1. Clone the repo: git clone https://github.com/sourcegraph/sourcegraph/ 
2. Make sure your node environment is running version 16.x.x.
3. Please sign our [contributor license agreement](https://docs.google.com/forms/d/1Z8zQHZs1ycfOCaR8N43S2lalgsmTTeKoUizrc93xk6M/edit?usp=drive_web). Once you have already signed it, you will have permission to merge your contributions after an appropriate review and approval. 
If one or more people ask for the same issue, the first person asking for it will have the priority for the assignation.
4. If you have any questions, please [refer to the docs first](https://docs.sourcegraph.com/). If you don’t find any relevant information, mention the issue author.
5. Issue author will try to provide guidance. Sourcegraph always works in async mode. We will try to answer as soon as possible, but please keep time zones differences in mind.

## Relevant development docs

### Getting applications up and running

- [Developing the Sourcegraph web app](https://docs.sourcegraph.com/dev/background-information/web/web_app#commands)
- [Table of contents](https://docs.sourcegraph.com/dev/background-information/web)
- Configuring backend services locally is not required for most frontend issues. However, a guide on how to do this can be found here.
### How to style UI
- [Guidelines](https://docs.sourcegraph.com/dev/background-information/web/styling)
- [Wildcard Component Library](https://docs.sourcegraph.com/dev/background-information/web/wildcard)
  ### Client packages [overview](https://github.com/sourcegraph/sourcegraph/blob/main/client/README.md)
### How to write tests
  - [testing web code](https://docs.sourcegraph.com/dev/background-information/testing_web_code)
  - [testing principles](https://docs.sourcegraph.com/dev/background-information/testing_principles)
  - [how to testing](https://docs.sourcegraph.com/dev/how-to/testing)
### Continuous integration pipeline
- [Github actions](https://github.com/sourcegraph/sourcegraph/actions)
- [Visual testing](https://docs.sourcegraph.com/dev/background-information/testing_principles#visual-testing)
  - We use percy.io and chromatic.com for visual testing.
- [Buildkite](https://buildkite.com/sourcegraph/sourcegraph) is used to run most of our tests.
### Pull Requests
- [How to structure](https://docs.sourcegraph.com/dev/background-information/code_reviews#what-makes-an-effective-pull-request-pr)
- [Size guidelines](https://about.sourcegraph.com/handbook/engineering/developer-insights#prefer-small-prs-lines)
- Git branch name convention: `[developer-initials]/short-feature-description`
- [Examples on Github](https://github.com/sourcegraph/sourcegraph/pulls?q=is%3Apr+label%3Ateam%2Ffrontend-platform)


