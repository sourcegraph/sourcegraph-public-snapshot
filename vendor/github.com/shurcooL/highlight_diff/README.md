highlight_diff
==============

[![Build Status](https://travis-ci.org/shurcooL/highlight_diff.svg?branch=master)](https://travis-ci.org/shurcooL/highlight_diff) [![GoDoc](https://godoc.org/github.com/shurcooL/highlight_diff?status.svg)](https://godoc.org/github.com/shurcooL/highlight_diff)

Package highlight_diff provides syntaxhighlight.Printer and syntaxhighlight.Annotator implementations
for diff format. It implements intra-block character-level inner diff highlighting.

It uses GitHub Flavored Markdown .css class names "gi", "gd", "gu", "gh" for outer blocks,
"x" for inner emphasis blocks.

Installation
------------

```bash
go get -u github.com/shurcooL/highlight_diff
```

License
-------

-	[MIT License](https://opensource.org/licenses/mit-license.php)
