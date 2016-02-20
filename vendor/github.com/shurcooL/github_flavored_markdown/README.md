# github_flavored_markdown [![Build Status](https://travis-ci.org/shurcooL/github_flavored_markdown.svg?branch=master)](https://travis-ci.org/shurcooL/github_flavored_markdown) [![GoDoc](https://godoc.org/github.com/shurcooL/github_flavored_markdown?status.svg)](https://godoc.org/github.com/shurcooL/github_flavored_markdown)

Package github_flavored_markdown provides a GitHub Flavored Markdown renderer
with fenced code block highlighting, clickable header anchor links.

The functionality should be equivalent to the GitHub Markdown API endpoint specified at
https://developer.github.com/v3/markdown/#render-a-markdown-document-in-raw-mode, except
the rendering is performed locally.

See examples for how to generate a complete HTML page, including CSS styles.

Installation
------------

```bash
go get -u github.com/shurcooL/github_flavored_markdown
```

License
-------

- [MIT License](http://opensource.org/licenses/mit-license.php)
