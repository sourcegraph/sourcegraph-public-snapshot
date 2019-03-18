# Experimental language servers

Sourcegraph has experimental code intelligence support for Bash, Clojure, C++, C#, CSS, Dockerfile, Elixir, HTML, Lua, OCaml, R, Ruby, and Rust.

# Using an experimental language server on Sourcegraph Server

To use an experimental language server on Sourcegraph Server, open the code intelligence admin page in the site admin area (e.g., https://sourcegraph.example.com/site-admin/code-intelligence) and click **Enable**:

<img src="img/experimental-language-server-enable.png"/>

This will pull the language server Docker image from the [Sourcegraph repository](https://hub.docker.com/r/sourcegraph/) and run it. Check out the [lsp-adapter](https://github.com/sourcegraph/lsp-adapter) repository to see what's in the images and how Sourcegraph communicates with the language server.

# Caveats of experimental language servers

Language servers are built and maintained by many participants in the developer community. Consult each language server's documentation for information on its stage of development, level of support for LSP functionality, security guarantees, etc.

## Cross-repository definitions and references

Cross-repository definitions and references are part of an [extension](https://github.com/sourcegraph/language-server-protocol/blob/ba96cf4d529f1a5cd9ff227db5a3883651f95bcb/extension-symbol-descriptor.md) to the standard Language Server Protocol that not all language servers support (the **Additional capabilities** column on [langserver.org](http://langserver.org/) lists the ones that do). If you’d like to add support for these features to a language server, check out the [`extension-symbol-descriptor`](https://github.com/sourcegraph/language-server-protocol/blob/ba96cf4d529f1a5cd9ff227db5a3883651f95bcb/extension-symbol-descriptor.md) specification.

## External dependencies

Accurate and complete code intelligence often requires access to external depdendencies. Language servers that do not fetch and analyze external dependencies will likely only be able to provide limited code intelligence.

## Large repositories

All of the experimental language server images use [lsp-adapter](https://github.com/sourcegraph/lsp-adapter) to act as a translator between Sourcegraph and the language server. Unless the language server supports [`extension-files`](https://github.com/sourcegraph/language-server-protocol/blob/ba96cf4d529f1a5cd9ff227db5a3883651f95bcb/extension-files.md), lsp-adapter clones every repository into a clean directory for the language server to operate on, which increases the disk usage and slows down the initialization of large repositories.

## Isolation

If a language server is capable of executing arbitrary code (running build scripts, etc.), then you should be aware of the security risks of running this language server on your code. However, this risk is similar to that of a developer manually running this language server inside of their editor.

## Early-stage language servers

Many of these experimental language servers are still under active development. If you notice an issue, try checking out the language server's README or issue tracker (available from the code intelligence admin panel, the code intelligence status indicator on the file view, the [Sourcegraph repository on Docker Hub](https://hub.docker.com/r/sourcegraph/), or in the [lsp-adapter repository's Dockerfiles](https://github.com/sourcegraph/lsp-adapter/tree/876b1f35cf43e210a8b0e1623e19b0c3be73f7e7/dockerfiles)) for possible solutions. If you think that the issue is with Sourcegraph itself, please file it on our public [issue tracker](http://github.com/sourcegraph/issues). We’ll often contribute any fixes upstream, if appropriate.
