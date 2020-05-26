# Deeper dive into Sourcegraph search modes

> NOTE: If you're new to code search in general, check out our blog post [The find-and-replace Odyssey, a programmer's guide](https://about.sourcegraph.com/blog/a-programmers-guide-to-find-and-replace) which takes you on a journey from using basic search and replace in your editor to the most advanced commandline tools, and of course, Sourcegraph.

Sourcegraph provides three search modes with different syntax:

- **Exact string matching (literal)**<br/>
The default and simplest search mode.<br/>

- **Regular expressions (regexp)**<br/>
More complex, but awesome when you need to match based on patterns instead of being limited to an exact string.<br/>

- **Structural search**<br/>
A brand new type of language aware search that understands the syntactic structure of programming languages.

Learn more about each search mode and when to use them by watching the videos in each section below.

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

