# Improving developer onboarding

When new developers join a team, it's important to help them understand the code base quickly because:

- Their effectiveness and happiness depends on it. (This is obvious.)
- They're most at risk of introducing bugs while they are getting up to speed. (This factor is underappreciated.)

# TODO!(sqs): below here is not yet updated. copied from code_review.md.

---














. Their effectiveness and happiness depends on it, and  Code review is a fundamental process of any development team. It's important for improving quality, sharing knowledge, and avoiding mistakes.

No matter what code review tool you use ([GitHub](https://github.com), [GitLab](https://gitlab.com), [Bitbucket](https://bitbucket.org), [Phabricator](https://phabricator.org), or [something else](TODO!(sqs): link to code review tools list)), the effectiveness of code review depends on human factors:

- Does the reviewer understand the changed code?
- Is it clear how the change change affects the entire system (including other services and/or repositories)?
<!-- - Are the right developers reviewing the code? TODO(sqs): This involves saved searches. It can be added to this doc later. -->

This document explains how [Sourcegraph](https://sourcegraph.com) helps with these important parts of your code review process. Sourcegraph is open-source and self-hosted, and it integrates with all the code review tools listed above. You can [start using it](../../../index.md#quickstart) in a few minutes for your organization.

> Want us to highlight how Sourcegraph helps your organization with code reviews? [Submit a PR](TODO!(sqs): add link) to add your experiences to this document. Many companies (including those with 1,000+ developers) use Sourcegraph in precisely the ways described here; we're working on engineering blog posts with them to link here, too.

## Help reviewers understand the changed code

### Scenario

You're asked to review several hundred lines of changed code. You open up the diff or pull request, and a red-green glow fills your office. It's a bunch of added and deleted code you need to review. You're familiar with the code, having written most of it a few months ago. But some of it has changed, and you're trying to refresh your memory of how it all works.

### Problem

To review changed code effectively, you need to actually understand the code:

- What calls what?
- Are the right arguments given?
- Are errors from function calls checked correctly?
- Etc.

<!-- TODO(sqs): I don't like how the bullets above turned out. Will revisit (suggestions appreciated). -->

You wouldn't *write* code without this information, so why would you *review* code without it? This information is usually available in editors (using go-to-definition, find-references, and hover tooltips), but none of the code review tools above offer it.

You could check out the branch for the code review locally and open the code in your editor to get this information, but in our experience, that's rarely done. If you're the rare, diligent developer who always does that, then we applaud you. But your teammates likely aren't so diligent, and wouldn't it be great to save a few minutes each time you do a review (and be able to add review comments without switching back and forth)?

### Solution

Sourcegraph integrates with your code review tool in a "feels-like-native" way to add code intelligence: go-to-definition, find-references, and hover tooltips. This makes it easy to navigate around code just as you would in your editor. Just hover over any token in the diff or pull request to see it in action:

TODO!(sqs): add gif

Want this? [Install Sourcegraph.](../../../index.md#quickstart)

## Help reviewers see how the change affects other services and/or repositories

### Scenario

You're asked to review a minor simplification to a core library inside your organization. It removes a function parameter that was deprecated and you think nobody is using anymore. If you're right, approving and merging the change will clean up some tech debt. If you're wrong, you'll confuse and block the other teams when they next upgrade their library dependencies.

### Problem

To review a change effectively, you need to know how it affects the entire system:

- What other services and repositories depend on the changed code?
- How do they call into the changed code?
- What teams and people rely on it?

It's not enough to just look at the diff in isolation, but that's how the code review tools above present the change.

### Solution

Sourcegraph supports cross-repository and cross-service go-to-definition and find-references for all of your organization's code. This means you can see repository and service dependencies down to individual method calls and arguments.

It works for both repository/package dependencies (using the package manager for your language) and RPC/service dependencies (using Protobuf, Thrift, Avro, Swagger, etc.).

For repository/package dependencies (TypeScript is shown here):

TODO!(sqs): insert gif

For service/RPC dependencies (Thrift is shown here):

TODO!(sqs): insert gif

Want this? [Install Sourcegraph.](../../../index.md#quickstart)

## Next steps

When a developer is reviewing code, they [deserve](https://about.sourcegraph.com/plan) to have the information mentioned in this document to ensure their time is well spent. If this information helps developers catch a few more issues in code review (or even enjoy reviewing code more), setting up Sourcegraph is worth it ([it's easy](../../../index.md#quickstart)).

To improve your organization's code reviews (without switching tools or processes), [install Sourcegraph](../../../index.md#quickstart).
