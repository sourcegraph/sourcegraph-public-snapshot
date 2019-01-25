# Tracking issue template

This is the template for the [project tracking issue](index.md#planning) that the [roadmap](../roadmap/index.md) links to.

1. [Create a new GitHub issue on sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph/issues/new?label=roadmap)
   - Label: `roadmap` and any other relevant labels
   - Milestone: the milestone of the [release](../releases.md) that includes this project. If a project spans multiple releases, then break it into multiple tracking issues (one per release).
1. Use the project name as the issue title.
1. Paste in the template below and fill it in.

```markdown
<!-- A short description of the problems this plan addresses and what will ship in its associated release milestone. -->

## Plan

<!-- Either a Markdown checklist of tasks with issue links as needed: -->

- [ ] Issue 1 title #123
- [ ] Issue 2 title #456

<!-- or a link to an issue query: -->

[Issues](https://github.com/sourcegraph/sourcegraph/issues?q=is:issue+is:open+sort:updated-desc+label:mylabel+milestone:3.1)

## Test/review plan

<!-- This part of the template is experimental. We'll see how it goes for 3.1 planning. -->

- Code reviewer: <!-- fill in @user(s) -->
- Tester: <!-- fill in @user(s) -->

<!-- This is how the project will be tested for the release. Fill this out with at least high-level details right now, and finish it by one week before the release. -->

<!-- Add other sections if needed to describe: future project work planned for the next release, technical or deadline risks, blockers/dependencies, backcompat/migration, or anything else important. -->
```

