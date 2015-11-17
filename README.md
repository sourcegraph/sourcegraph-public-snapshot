# Sourcegraph: the intelligent and hackable code platform

Sourcegraph makes your team more collaborative and efficient by
letting them:

* search, browse, and cross-reference code (like an IDE)
* view live usage examples for any function, type, etc.
* perform code reviews
* carry on persistent discussions on any piece of code
* discover important information in context
* track issues and projects

Status: **limited release** ([Go](https://golang.org) and [Java](http://docs.oracle.com/javase/8/)
support only)

Your git repositories can live on Sourcegraph, or you can use it to
search and browse existing repositories.

## Installation

See the [Sourcegraph Docs](https://src.sourcegraph.com/sourcegraph/.docs)
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
[CONTRIBUTING.md](CONTRIBUTING.md). We welcome all types of
contributions--code, documentation, assets, community support, and
user feedback.

Our [README.dev.md](README.dev.md) is a good place to start.

## Security

Security is very important to us. If you discover a security-related
issue, please responsibly disclose it by emailing
[security@sourcegraph.com](mailto:security@sourcegaph.com) and not by
creating an issue.

[Read our complete security policy.](https://sourcegraph.com/security)

## License

[Fair Source License](http://fair.io)
