# Sourcegraph: the intelligent and hackable code host

Sourcegraph is a self-hosted Git repository service with Code
Intelligence. It runs on your own server or cloud, takes 10 minutes to
install, and gives your team the power to build better software.

* Sourcegraph's Code Intelligence analyzes and understands your code,
  letting you browse code like an IDE.
* Live usage examples save tons of time and help spread best practices.
* Smart code search quickly gets you what you need.
* You can start discussions and create issues right inline with your
  code and have them stay attached even when the code changes.

Watch a [demo video](https://www.youtube.com/watch?v=XOdh3-QJSzs) and
see the
[announcement blog post](https://sourcegraph.com/blog/133554180524/announcing-the-sourcegraph-developer-release-an).

Status: **limited release** ([Go](https://golang.org) and [Java](http://docs.oracle.com/javase/8/)
support only)

## Installation

See the [Sourcegraph docs](https://src.sourcegraph.com/sourcegraph/.docs)
for installation instructions.

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
[security@sourcegraph.com](mailto:security@sourcegaph.com) and not by
creating an issue.

[Read our complete security policy.](https://sourcegraph.com/security)

## License

[Fair Source License](https://fair.io)
