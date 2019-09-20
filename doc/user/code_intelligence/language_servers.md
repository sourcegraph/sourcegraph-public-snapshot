# Language servers

> Language servers are not recommended because they are complex to configure, require separate deployment, and are slow to initialize. Instead, we recommend using [LSIF](./lsif.md) to get precise code intelligence.

Language servers provide more precise code intelligence than the out-of-the-box experience. There are language servers for the following languages:

- [Go](https://sourcegraph.com/extensions/sourcegraph/go)
- [TypeScript](https://sourcegraph.com/extensions/sourcegraph/typescript)
- [Python](https://sourcegraph.com/extensions/sourcegraph/python)
- More can be found in the [extension registry](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22)

Language servers run securely in your self-hosted Sourcegraph instance. The extensions (and associated language servers) perform advanced, scalable code analysis and are derived from our popular open-source language servers in use by hundreds of thousands of developers in editors and on Sourcegraph.com.

## Language server deployment

Most Sourcegraph extensions that provide code intelligence require a server component, called a language server. These language servers are usually deployed alongside other Sourcegraph services in another Docker container or within the same Kubernetes cluster. Check the corresponding extension documentation for deployment instructions.

### Open standards

Code intelligence is powered by [Sourcegraph extensions](../index.md) and language servers based on the open-standard Language Server Protocol (published by Microsoft, with participation from Facebook, Google, Sourcegraph, GitHub, RedHat, Twitter, Salesforce, Eclipse, and others).

Hundreds of thousands of developers already use Sourcegraph's language servers in their editor or while browsing public code on [Sourcegraph.com](https://sourcegraph.com). Microsoft's [Visual Studio Code](https://code.visualstudio.com) and GitHub's [Atom](https://atom.io) editors both use Sourcegraph language servers in official editor extensions. The language servers used for code intelligence on Sourcegraph are based on our widely used language servers, with extensive improvements for performance, cross-repository definitions and references, security, isolation, type/build inference, and robustness.

For more information about the Language Server Protocol (LSP), visit [Microsoft's official LSP site](https://microsoft.github.io/language-server-protocol/). For a more detailed list of existing language servers, visit [langserver.org](https://langserver.org) (maintained by Sourcegraph).
