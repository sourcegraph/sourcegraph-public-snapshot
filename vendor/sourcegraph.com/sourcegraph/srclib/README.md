# srclib [![Build Status](https://travis-ci.org/sourcegraph/srclib.png?branch=master)](https://travis-ci.org/sourcegraph/srclib) 

*Note: srclib is alpha.
[Post an issue](https://github.com/sourcegraph/srclib/issues) if you have any
questions or difficulties running and hacking on it.*

[**srclib**](https://srclib.org) is a source code analysis library. It
provides standardized tools, interfaces and data formats for
generating, representing and querying information about source code in
software projects.

**Why?** Right now, most people write code in editors that don't give them as
much programming assistance as is possible. That's because creating an editor
plugin and language analyzer for your favorite language and editor combo is a
lot of work. And when you're done, your plugin only supports a single language
and editor, and maybe only half the features you wanted (such as doc lookups and
"find usages"). Because there are no standard cross-language and cross-editor
APIs and formats, it is difficult to reuse your plugin for other languages or
editors.

We call this the **M-by-N problem**: given *M* editors and *N* languages, we
need to write (on the order of) *M*&times;*N* plugins to get good tooling in
every case. That number gets large quickly, and it's why we suffer from poor
developer tools.

srclib solves this problem in 2 ways by:

* Publishing standard formats and APIs for
  source analyzers and editor plugins to use. This means that improvements in a
  srclib language analyzer benefit users in any editor, and improvements in a
  srclib editor plugin benefit everyone who uses that editor on any language.

* Providing high-quality language analyzers that implement this
  standard. These power [Sourcegraph.com](https://sourcegraph.com).

Step 1 is to build the language analyzers. Next up, we will build the
editor plugins!

See [srclib.org](https://srclib.org) for more information.

Currently, srclib supports:

* **Languages:** [Go](https://sourcegraph.com/sourcegraph/srclib-go), [JavaScript](https://github.com/sourcegraph/srclib-javascript), and [Ruby](https://github.com/sourcegraph/srclib-ruby) (coming very soon: [Python](https://sourcegraph.com/sourcegraph/srclib-python), [Java](https://github.com/sourcegraph/srclib-java), and [Bash](https://github.com/sourcegraph/srclib-bash))

* **Features:** jump-to-definition, find usages, type inference, documentation
  generation, and dependency resolution

Want to extend srclib to support more languages, features, or editors?
We'll help!
[Post an issue](https://github.com/sourcegraph/srclib/issues) to let
us know what you're building to get started.


## Usage

See [*Getting started*](https://srclib.org/#install) for installation
instructions.

# Misc.

* **bash completion** for `srclib`: run `source contrib/completion/srclib-completion.bash` or
  copy that file to `/etc/bash_completion.d/srclib` (path may be different
  on your system)

# Development

## srclib binary release process

Contributors with deploy privileges can update the official binaries
via these instructions:

1. `go install github.com/laxer/goxc`
1. Ensure you have the AWS credentials set so that the AWS CLI (`aws`) can write to the `srclib-release` S3 bucket.
1. Run `make release V=1.2.3`, where `1.2.3` is the version you want to release (which can be arbitrarily chosen but should be the next sequential git release tag for official releases).


## License
srclib is licensed under the [MIT License](https://tldrlegal.com/license/mit-license).
More information in the LICENSE file.


## Contributing

**If you want to start hacking on srclib or write your own srclib toolchain, [join the srclib Slack](http://slackin.srclib.org) and then access it on [srclib.slack.com](https://srclib.slack.com).**
Don't forget to introduce yourself on the #General channel. We are more than happy to meet new contributors and to help people to get started.
