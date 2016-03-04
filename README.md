# Sourcegraph: Code Intelligence for teams

Sourcegraph is a **Git repository service** with Code Intelligence. It
**runs on your own server** or cloud and installs in 5 minutes, or you
can [use Sourcegraph on Sourcegraph.com](https://sourcegraph.com/join) for free.

Sourcegraph gives your team the power to build better software with:

* **Code Intelligence:** Understand code more quickly with language-aware jump-to-definition and tooltips. (Go and Java only, more languages coming soon.)
* **Live usage examples:** See how any function, class, etc., is currently being used across your codebases. As a wise developer once said, "The right example is worth a thousand words of documentation."
* **Better code reviews:** Review changesets more effectively with drafts and Code Intelligence context in diffs---and an easy branch-based pull request model.
* **Smart search:** Find code quickly by function name, full text, etc.
* **Deep integrations:** Works great standalone or with GitHub, GitHub Enterprise, JIRA, and more.
* **Hackable source code:** [Sourcegraph's source code](https://src.sourcegraph.com/sourcegraph) is publicly available under the [Fair Source License](https://fair.io).

[**Get started with your own Sourcegraph server**](https://src.sourcegraph.com/sourcegraph/.docs/getting-started/) in 5 minutes! Want to try it out first? You're on a Sourcegraph server ([src.sourcegraph.com](https://src.sourcegraph.com)) now, so just browse around this server.

*More info? Watch the [demo video](https://www.youtube.com/watch?v=XOdh3-QJSzs),
see the
[announcement blog post](https://sourcegraph.com/blog/133554180524/announcing-the-sourcegraph-developer-release-an),
and [view enterprise capabilities](https://sourcegraph.com).*


## Installation

Follow the 5-minute
[Sourcegraph installation instructions](https://src.sourcegraph.com/sourcegraph/.docs/getting-started/). For
more installation methods, check out the
[docs](https://src.sourcegraph.com/sourcegraph/.docs).


## Under the hood

Sourcegraph is built on several components:

* [srclib](https://srclib.org), a multi-language, hackable source code
  analysis toolchain
* The [Go](http://golang.org) programming language
* [gRPC](http://grpc.io), an HTTP2-based RPC protocol that uses
  Protocol Buffer service definitions
* [React](https://facebook.github.io/react/), a JavaScript library for
  building UIs.
* [Sourcegraph.com](https://sourcegraph.com), a public instance of
  Sourcegraph that provides information about open-source projects to
  your local Sourcegraph.

## Contributing to Sourcegraph

Want to make Sourcegraph better? Great! Check out
[CONTRIBUTING.md](https://src.sourcegraph.com/sourcegraph@master/.tree/CONTRIBUTING.md). We
welcome all types of contributions--code, documentation, assets,
community support, and user feedback.

Our
[README.dev.md](https://src.sourcegraph.com/sourcegraph@master/.tree/README.dev.md)
is a good place to start.

## Security

Security is very important to us. If you discover a security-related
issue, please responsibly disclose it by emailing
[security@sourcegraph.com](mailto:security@sourcegraph.com) and not by
creating an issue.

[Read our complete security policy.](https://sourcegraph.com/security)

## License

[Fair Source License](https://fair.io)
