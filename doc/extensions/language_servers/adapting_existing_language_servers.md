# Adapting an existing language server for use with Sourcegraph'

This documentation page is intended for language server developers who are adapting an existing language server (or building a new language server) to provide code intelligence on Sourcegraph.

Sourcegraph provides [code intelligence](index.md) by communicating with language servers that adhere to the Language Server Protocol standard, plus a few additional protocol extensions and requirements. These additional requirements are necessary because unlike in the common case where language servers are used for local code editing, language servers running inside Sourcegraph don't have access to the developer's existing repository checkout on their machine.

## Implementation requirements

With [lsp-adapter](https://github.com/sourcegraph/lsp-adapter), there are no Sourcegraph-specific requirements. Just pass the command to run your language server to `lsp-adapter`:

```
lsp-adapter -proxyAddresss=0.0.0.0:1234 mylang-exe
```

This will forward communication on port 1234 to the stdio of `mylang-exe`. Jump to the next section to [configure Sourcegraph](#Configuring-Sourcegraph-to-use-a-new-language-server).

Without lsp-adapter, a language server must support the following:

- [extension-files](https://github.com/sourcegraph/language-server-protocol/blob/master/extension-files.md): LSP extension for listing and retrieving files from the target repository.

- Non-`file:` root URIs: the language server must accept `initialize` requests with root URIs of any scheme (it lists and fetches the files using the extension-files LSP extension for non-`file:` URIs). As a convention, Sourcegraph uses `git://NAME?REV` as the root URI for a repository with the given name (e.g., `example.com/foo`) at the given revision (the 40-character Git commit SHA).

- Multiple independent sessions: the language server must support multiple independent and simultaneous LSP/JSON-RPC2 connections (instead of only supporting a single global session).

The following are optional:

- Automatic build system handling: the language server can process a repository's build system configuration (npm/Maven/Gradle/Bazel/Pants/etc.) and perform build tasks, such as fetching dependency files and running codegen. If this is not fully implemented, the language coverage (of hovers, definitions, references, etc.) will be incomplete.

- [extension-workspace-references](https://github.com/sourcegraph/language-server-protocol/blob/master/extension-workspace-references.md): LSP extension to enable cross-repository references.

- [extension-cache](https://github.com/sourcegraph/language-server-protocol/blob/master/extension-cache.md): LSP extension that provides the language server with a simple cache interface for implementation-specific caching.

## Configuring Sourcegraph to use a new language server

Sourcegraph communicates with language servers over TCP.

1.  Start your new language server and ensure it's listening for TCP connections on an address and port that is accessible to Sourcegraph. (We assume it's listening on `tcp://mylang:1234`.)

2.  Modify the Sourcegraph site configuration's "langservers" property as described in the [code intelligence documentation](install/index.md) to add a new entry for your language:

    ```json
    "langservers": [
      ...,
      {
        "language": "mylang",
        "address": "tcp://mylang:1234"
      },
      ...
    ]
    ```

    The name "mylang" should be the lowercase name of the language used by [Linguist](https://github.com/github/linguist/tree/master/samples), the language detection library that Sourcegraph uses.

3.  Start or restart Sourcegraph with the new site configuration.
4.  Browse to a file (written in the new language) inside a repository on Sourcegraph.
5.  Hover over a token to test that it's working. Also try clicking on the token and then going to its definition or finding references.

### Troubleshooting

- First, ensure that the language server works well on the same repository with a local editor, such as Visual Studio Code. To do so, you may need to create a shim editor extension.

  This is the best way to ensure that the language server is functioning correctly and to isolate the problem.

- Set the following environment variables in the `sourcegraph/server` Docker container (or the `lsp-proxy` deployment, for Sourcegraph cluster deployments to Kubernetes):

  ```
  LSP_PROXY_TRACE_FS_REQUESTS=1
  SRC_LOG_LEVEL=dbug
  ```

  Try hovering over tokens again and inspect the log output.

- To simplify testing, you can write tests that send a known-good sequence of LSP messages to Sourcegraph's `/.api/xlang` HTTP URL path and assert that the output matches an expected value.

  Tip: Perform a hover in your browser with devtools open and look for the `/.api/xlang/textDocument/hover` HTTP request. Right-click and choose "Copy as cURL" (in Chrome) to get a `curl` shell command that will perform the same request.
