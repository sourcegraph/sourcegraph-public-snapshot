# Contributing to Sourcegraph

Sourcegraph is [Fair Source licensed](https://fair.io) (not open
source). We welcome all types of contributions: bug reports, code,
documentation, and feedback.

This document outlines some of the conventions, resources, and contact
points for developers to make getting your contribution into
Sourcegraph easier.

### Community and contact information

- [Issue tracker](https://github.com/sourcegraph/sourcegraph/issues)
- [Pull requests](https://github.com/sourcegraph/sourcegraph/pulls)
- Email: [support@sourcegraph.com](mailto:support@sourcegraph.com)

## Getting started

* Use [Sourcegraph.com](https://sourcegraph.com) to get familiar with
  the product.
* Fork the
  [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph)
  repository on GitHub.
* See [docs/dev.md](./docs/dev.md) for build and test instructions.
* Submit bugs and patches!

## Reporting bugs and creating issues

You can report bugs in the
[Sourcegraph issue tracker](https://github.com/sourcegraph/sourcegraph/issues).

Please include the following information in your bug reports, if
possible:

* short title and summary of the issue
* steps to reproduce the issue, including URLs or commands to run
* the expected behavior and the actual behavior
* log output
* the versions of your OS, browser, etc.
* screenshots

## Submitting patches

We accept patches submitted as
[pull requests](https://github.com/sourcegraph/sourcegraph/pulls) on
GitHub.

To submit a pull request:

* Create a topic branch on your fork of Sourcegraph, usually based on
  `master`.
* Make commits, with each one being a logical unit of work.
* Push your branch to your fork and
  [submit a pull request](https://github.com/sourcegraph/sourcegraph/pulls)
  to the upstream Sourcegraph repository.
* Sign the [Sourcegraph CLA](./dev/CLA.txt) (contributor license
  agreement) and send it to us so we can incorporate your
  contributions.
* We'll review the change and merge it if it looks good.

If you're not sure whether we'd accept a change, you can
[file an issue](https://github.com/sourcegraph/sourcegraph/issues) to
discuss it beforehand.

We will compensate you for certain contributions, if you receive
explicit preapproval over email from an authorized individual at
Sourcegraph. Contact
[contributing@sourcegraph.com](mailto:contributing@sourcegraph.com) to
learn more and get started. (We will post a list of open projects here
soon.)

### Code style

See our [docs/style.md](docs/style.md) for more information.

### Commit message format

We do not currently follow a strict convention for commit
messages. However, we do generally try to provide commit messages
which answer two questions: what changed and why. The subject line
should feature the what and the body of the commit should describe the
why.

```
notif: add a Slack integration

You must configure a webhook URL via CLI flag, env, or config
for Sourcegraph to send Slack notifications about changeset activity.

Closes #38
```

The format can be described more formally as follows:

```
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
```

The first line is the subject and should be no longer than 70
characters, the second line is always blank, and other lines should be
wrapped at 80 characters.  This allows the message to be easier to
read in various git tools.
