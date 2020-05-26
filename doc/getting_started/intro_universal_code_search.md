# Introduction to Sourcegraph Universal Code Search

At this point, you've [chosen the right deployment model](choosing_the_right_deployment_model.md), [installed Sourcegraph](installing_sourcegraph.md), [connected your code host](connecting_your_code_host.md), and now, you're ready to begin using Sourcegraph Universal Code Search.

If you haven't used a code search tool before, it may not be obvious how to get started, so first, we'll explain what Universal Code Search is, then move onto getting you comfortable using Sourcegraph with an in-depth tour for first-time users.

## What is Sourcegraph Universal Code Search?

Universal Code Search provides capabilities beyond what your editor and code host can provide and is unique to Sourcegraph. It has a conceptual meaning, as well as a concrete one.

Conceptually, Sourcegraph is your vehicle for searching, exploring, and navigating across your entire universe of code, meaning every codebase in your organization. Code search must be universal in order to be effective, as it's no longer feasible for individual developers to download and search all code locally. While your code host provides code search, it's extremely limited and not universal, as search is limited to the repositories on that code host only.

In concrete terms, Universal Code Search is the ability to provide access to your universe of code by:

- Providing [exact string (literal)](../user/search/queries.md#literal-search-default), [regexp](../user/search/queries.md#regexp-search), and (new) [structural search](../user/search/queries.md#structural-search) query syntax
- Supporting [every Git](../admin/external_service.md) and [non-Git](../admin/external_service/non-git.md) code host through either native integrations or custom solutions
- Searching in every repository from multiple code hosts simultaneously
- Searching in any repository branch
- Searching not just code, but commit diffs, and commit messages
- Providing fast and precise [code intelligence for every popular language](../user/code_intelligence.md)
- Providing [integrations](../integration/index.md) for everywhere you read and write code (editors, IDEs, code hosts)
- Providing a variety of [deployment options](../admin/install/index.md) to operate at massive scale, e.g., 40,000+ repositories
- Integrating [data and insights from external developer tools](https://sourcegraph.com/extensions?query=category%3A%22External+services%22) when reviewing code, such as [code coverage overlays](https://sourcegraph.com/extensions/sourcegraph/codecov)

Now that you know what Universal Code Search is, let's explore how to use it by taking you on a tour of the Sourcegraph UI.

## Introduction to using Sourcegraph

This video will show you to how to build a search query step-by-step, as well as customizing the Sourcegraph UI with custom search filters and [saving frequently used searches](../user/search/saved_searches.md).

<div class="container my-4 video-embed embed-responsive embed-responsive-16by9">
    <iframe class="embed-responsive-item" src="https://www.youtube.com/embed/D2x037j3BZ4?autoplay=0&amp;cc_load_policy=0&amp;start=0&amp;end=0&amp;loop=0&amp;controls=1&amp;modestbranding=0&amp;rel=0" allowfullscreen="" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" frameborder="0"></iframe>
</div>

Now that you're well acquainted with the Sourcegraph UI and constructing search queries, let's dive deeper into the different search modes to understand when you should use one over the other.

[**» Next: Deeper dive into Sourcegraph search modes**](deeper_dive_search_modes.md)
