# Using Universal Code Search

Universal Code Search has a conceptual meaning, as well as a concrete one.

Conceptually, Sourcegraph is your vehicle for searching, exploring, and navigating across your entire universe of code, meaning every codebase in your organization. Code search must be universal in order to be effective, as it's no longer feasible for individual developers to download and search all code locally. While this is something a code host can do, it is also not universal, as search is limited to that code host only.

In concrete terms, Universal Code Search is our ability to provide access to this universe of code by:

- Provide exact string, regular expression, and structural search (new) query syntax
- Supporting every Git and non-Git code host through either native integrations or custom solutions
- Searching across every repository from multiple code hosts simultaneously
- Searching in every repository, on every branch
- Searching not just code, but commit diffs, and commit messages
- Providing fast and precise code intelligence for every popular language
- Providing integrations for everywhere you read and write code (editors, IDEs, code hosts)
- Providing a variety of deployment options to operate at massive scale, e.g., 40,000+ repositories
- Integrating data and insights from third party systems such as code coverage tools into code reviews

Now that you know what Universal Code Search is, let's now explore how to use it.

> NOTE: This section is changing rapidly with new how-to search videos being added weekly, so be sure to check soon to see what's new!

## The Sourcegraph search query modes

Sourcegraph provides three types of search query syntax:

- **Exact string matching (literal)**<br/>
The simplest to get started with and is the default search mode<br/>

- **Regular expressions (regexp)**<br/>
More complex, but awesome when you need to match based on patterns instead of being limited to an exact string<br/>
- **Structural search**<br/>
A brand new type of language aware search that understands the syntactic structure of languages.

We recommend starting with literal (exact string) searches, but so you know how each of the work, check-out the below videos which provide simple explanations with a real example for each search mode.

## Exact string search (literal search)

The "literal search" mode is the default when first using Sourcegraph, and in this screencast, you'll learn when to use it, and how using the `content` filter can help you take your literal searches one step further.

<div style="padding:56.25% 0 0 0;position:relative;">
    <iframe src="https://www.youtube.com/embed/CX6F5oCjfoc" style="position:absolute;top:0;left:0;width:100%;height:100%;" frameborder="0" webkitallowfullscreen="" mozallowfullscreen="" allowfullscreen=""></iframe>
</div>

## Regular expression search

Exact string (literal) searches are great for simple cases, but regular expressions in search queries and filters (such as `file`), amplify your ability to discover code in more places.

<div style="padding:56.25% 0 0 0;position:relative;">
    <iframe src="https://www.youtube.com/embed/J9k7l5W1qbk" style="position:absolute;top:0;left:0;width:100%;height:100%;" frameborder="0" webkitallowfullscreen="" mozallowfullscreen="" allowfullscreen=""></iframe>
</div>

## Structural search

Sourcegraph structural search is a code-aware search syntax that when enabled, lets you match nested expressions and whole code blocks that can be difficult or awkward to match using regular expressions.

Learn more by checking out the [structural search docs](../user/search/structural.md) and the [going beyond regular expressions with structural code search blog post](https://about.sourcegraph.com/blog/going-beyond-regular-expressions-with-structural-code-search/).

<div style="padding:56.25% 0 0 0;position:relative;">
    <iframe src="https://www.youtube.com/embed/Lg4cYEoSHeo" style="position:absolute;top:0;left:0;width:100%;height:100%;" frameborder="0" webkitallowfullscreen="" mozallowfullscreen="" allowfullscreen=""></iframe>
</div>

---

## Diving deeper into Universal Code Search



[**» Next: Universal Code Intelligence**](universal_code_intelligence.md)
