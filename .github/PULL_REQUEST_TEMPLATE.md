<!--

💡 To write a useful PR description, make sure that your description covers:

- WHAT this PR is changing:
    - How was it PREVIOUSLY.
    - How it will be from NOW on.
- WHY this PR is needed.
- CONTEXT, i.e. to which initiative, project or RFC it belongs.

The structure of the description doesn't matter as much as covering these points, so use
your best judgement based on your context.

Learn how to write good pull request description: https://www.notion.so/sourcegraph/Write-a-good-pull-request-description-610a7fd3e613496eb76f450db5a49b6e?pvs=4
-->

## Changelog

<!--
1. Ensure your pull request title is formatted as: $type($domain): $what
2. Add bullet list items for each additional detail you want to cover (see example below)
3. You can edit this after the pull request was merged, as long as release shipping it hasn't been promoted to the public.
4. For more information, please see this how-to https://www.notion.so/sourcegraph/Writing-a-changelog-entry-dd997f411d524caabf0d8d38a24a878c?

Cheat sheet: $type = chore|fix|feature $domain: source|search|ci|release|plg|cody|...
-->

<!--
Example:

Title: fix(search): parse quotes with the appropriate context
Changelog section:

## Changelog

- When a quote is used with regexp pattern type, then ...
- Refactored underlying code.
-->

## Test plan

<!-- All pull requests REQUIRE a test plan: https://docs.sourcegraph.com/dev/background-information/testing_principles

Why does it matter?

These test plans are there to demonstrate that are following industry standards which are important or critical for our customers.
They might be read by customers or an auditor. There are meant be simple and easy to read. Simply explain what you did to ensure
your changes are correct!

Here are a non exhaustive list of test plan examples to help you:

- Making changes on a given feature or component:
  - "Covered by existing tests" or "CI" for the shortest possible plan if there is zero ambiguity
  - "Added new tests"
  - "Manually tested" (if non trivial, share some output, logs, or screenshot)
- Updating docs:
  - "previewed locally"
  - share a screenshot if you want to be thorough
- Updating deps, that would typically fail immediately in CI if incorrect
  - "CI"
  - "locally tested"
-->
