+++
title = "Toolchain Troubleshooting"
navtitle = "Toolchains"
+++

Sourcegraph uses [srclib](https://srclib.org) toolchains to parse and perform analysis on your source code at the AST level, here we'll discuss common questions and problems relating to toolchains.

# Why Toolchains Matter

Toolchains allow Sourcegraph to perform intelligent features like clicking on a symbol to see where it's defined, or to find usage examples of any function in your entire codebase. Without an appropriate toolchain installed for your language, the feature set that Sourcegraph can provide to you is degraded significantly, so it's very important that you install a toolchain for your language.

# Installing toolchains

Sourcegraph currently supports both Go and Java, to install a toolchain on your Sourcegraph server just run:

* Mac OS X: `src toolchain install <language>`
* Linux: `sudo -u sourcegraph -i src toolchain install <language>`

For example `sudo -u sourcegraph -i src toolchain install go` to install the Go toolchain on Linux.

# Adding support for your language

If we don't yet support your language, there are two options:

1. **Contribute support for your language.**
  * Effectively you would [write and contribute a srclib toolchain](https://sourcegraph.com/github.com/sourcegraph/srclib) for your language. We'd be delighted to help you through this process in case you have questions!
1. **Vote for your favorite language.**
  * Send an email to [support@sourcegraph.com](mailto:support@sourcegraph.com) with the title "Vote: &lt;my language&gt;" and we'll try our best to add support for the most desired languages!

Note that even if we don't support your language, _you can still continue using Sourcegraph as a system of record_ -- just some core features like jump-to-defition, usage examples, etc. will be disabled.
