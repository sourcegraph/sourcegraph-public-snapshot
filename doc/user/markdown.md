---
ignoreDisconnectedPageCheck: true
---

# Sourcegraph-flavored Markdown

Sourcegraph uses [GitHub Flavored Markdown (GFM)](https://github.github.com/gfm/) anywhere Markdown is rendered. There are two different ways rendering happens:

1. Any **Markdown files in your repositories** can be viewed as the raw file, or formatted. The formatted view will be rendered using [Marked.js](https://marked.js.org/#/README.md#README.md), a JavaScript Markdown renderer, configured to use GFM.
2. **Sourcegraph code discussions** are rendered on the server using [github_flavored_markdown](https://godoc.org/github.com/shurcooL/github_flavored_markdown), a Go Markdown renderer.

If at any point you encounter rendering issues that do not meet GFM standards, this is unexpected, so please let us know.
