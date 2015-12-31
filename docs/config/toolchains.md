+++
title = "Language support"
description = "Enable Code Intelligence for your programming language"
+++

Sourcegraph uses [srclib](https://srclib.org) toolchains to parse and perform analysis on your source code at the AST level. Here we'll discuss common questions and problems relating to toolchains.

# Why language support matters

Toolchains allow Sourcegraph to deliver features like clicking on a symbol to see where it's defined or finding usage examples of any function in your codebase. Without a toolchain installed for your language, Sourcegraph features degrade significantly so it's important that you install language support.

# Installing toolchains

Sourcegraph currently supports both Go and Java. To install a toolchain on your Sourcegraph server just run:

* **Mac OS X:** `src srclib toolchain install <language>`
* **Linux:** `sudo -u sourcegraph -i src srclib toolchain install <language>`

Where `<language>` is either `go` or `java`.

# Adding support for your language

If we don't yet support your language, there are two options:

1. **Contribute support for your language.**
  * [Write and contribute a srclib toolchain](https://sourcegraph.com/github.com/sourcegraph/srclib) for your language. We'd be delighted to help you through this process in case you have questions!
1. **Vote for your favorite language.**
  * Send an email to [help@sourcegraph.com](mailto:help@sourcegraph.com) with the title "Vote: &lt;my language&gt;"!

Note that even if we don't support your language, _you can still continue using Sourcegraph as a Git repository host_ -- just some core features like jump-to-definition, usage examples, etc. will be disabled.
