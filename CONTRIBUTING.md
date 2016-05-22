# How to contribute

Sourcegraph is [Fair Source licensed](https://fair.io) and accepts contributions.
This document outlines some of the conventions, resources, and contact points for
developers to make getting your contribution into Sourcegraph easier.

## Getting started

- Clone the repo from https://sourcegraph.com/sourcegraph/sourcegraph
- Read the docs/dev.md for build instructions

## Contacting us

- Email: [support@sourcegraph.com](mailto:support@sourcegraph.com)
- Open a public thread: see below

## Reporting bugs and creating issues

Reporting bugs is one of the best ways to contribute. However, a good bug report
has some very specific qualities, so please read over our short document on
[reporting bugs](https://sourcegraph.com/sourcegraph/sourcegraph@master/-/tree/docs/dev/bugs.md)
before you submit your bug report.

[Contact support](mailto:support@sourcegraph.com) when you're ready
to file an issue.

## Contribution flow

This is a rough outline of what a contributor's workflow looks like today:

- Create a topic branch from where you want to base your work. This is usually master.
- Make commits of logical units.
- Create a patch via [`git format-patch`](https://ariejan.net/2009/10/26/how-to-create-and-apply-a-patch-with-git/)
and send it to us [via email](mailto:contributing@sourcegraph.com).
- Sign the CLA (in misc/CLA.txt) and send it to us so we may incorporate your contributions.
- We'll open a changeset and have a conversation about the patch.
- We'll merge the changeset if everything looks good.

We are working on streamlining this process so you can submit changesets directly instead of via email.
Thanks for your contributions!

### Code style

See our [docs/style.md](docs/style.md) for more information.

### Commit message format

We do not currently follow a strict convention for commit messages. However, we do
generally try to provide commit messages which answer two questions: what changed
and why. The subject line should feature the what and the body of the commit should
describe the why.

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

The first line is the subject and should be no longer than 70 characters, the
second line is always blank, and other lines should be wrapped at 80 characters.
This allows the message to be easier to read in various git tools.
