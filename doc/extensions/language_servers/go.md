# Go: code intelligence configuration

This page describes additional configuration that may be needed for Go code intelligence on certain codebases. To enable Go code intelligence, see the [installation documentation](install/index.md).

---

## Debugging Go code intelligence

Additional configuration for Go code intelligence may be required in some cases:

- [Custom GOPATHs / Go monorepos](#custom-gopaths--go-monorepos)
- [Vanity import paths](#vanity-import-paths)

## Custom GOPATHs / Go monorepos

By default, Sourcegraph assumes that Go code in a repository represents Go packages that would be placed under `$GOPATH/src/...`. That is, a Go repository is assumed to only contain Go packages.

For some repositories, such as Go monorepos, this may not be the case. These repositories typically have an entire (or multiple) `$GOPATH` directories comitted to them, and the Go language server may not be able to provide code intelligence without being informed of this.

To inform Sourcegraph's Go language server that your repository contains an entire `$GOPATH` directory, you can use one of three options:

1.  **Auto-detection via `.vscode/settings.json`**

    Sourcegraph will automatically detect a Visual Studio Code `settings.json` file with a GOPATH configuration. You may already have one of these files if you are using Visual Studio Code with the Go extension. The file `.vscode/settings.json` would look like:

    ```json
    {
      "go.gopath": "${workspaceRoot}/YOUR_GOPATH"
    }
    ```

    In this case, Sourcegraph would look for a folder named `YOUR_GOPATH` in the root of the repository.

2.  **Auto-detection via `.envrc`**

    Sourcegraph will also automatically detect a GOPATH from an `.envrc` file in the root of the repository. You may already have one of these if you are using direnv. For example a file such as:

    ```bash
    export GOPATH=${PWD}/third_party
    GOPATH_add code:code2
    GOPATH_add /absolute
    ```

    Would lead to Sourcegraph using a final `GOPATH` of `third_party:code:code2`. Note that we will ignore any `/absolute` path, and that we do not execute `.envrc` files but rather scan them for simple syntax such as the above. If you use a more complex `.envrc` file to build your `GOPATH`, this auto-detection may not work for you.

3.  **Manual configuration via `.sourcegraph/config.json`**

    If you add a `.sourcegraph/config.json` file in the root directory of your repository, Sourcegraph will use this configuration to determine the `GOPATH` instead of the auto-detection methods described above. An example configuration is:

    ```json
    {
      "go": {
        "GOPATH": ["/third_party", "code"]
      }
    }
    ```

    Sourcegraph will use a final `GOPATH` of `third_party:code`. That is, it will assume the `third_party` and `code` directories in the root of the repository are to be used as `$GOPATH` directories.

---

## Vanity import paths

When the Go language server encounters a vanity import path, it must be able to locate the source code for it or else code intelligence will not work for code related to that dependency.

For example, consider a repository `github.com/example/server` which contains Go code with an `import "example.io/pkg/logger"` statement.

1.  If the source code for `example.io/pkg/logger` is located under a vendor directory, Sourcegraph will use that in the same manner that the `go` tool would.
2.  If the source code for `example.io/pkg/logger` is inside of the current repository at e.g. `github.com/example/pkg/logger`, Sourcegraph will look for it by scanning the repository for a [canonical import path comment](https://golang.org/doc/go1.4#canonicalimports) using some heuristics.

    - For example, if Sourcegraph finds a canonical import path comment such as `package logger // import "example.io/pkg/logger"` in the `pkg/logger` directory of the repository, Sourcegraph will assume that the code in the `pkg/logger` directory is what should be used when a `import "example.io/pkg/logger"` statement is seen.
    - Note that Sourcegraph only needs to find one such comment for the `example.io` domain in order to resolve all other vanity imports. That is, placing this comment in `pkg/logger/logger.go` is enough for Sourcegraph to know how to import any package under `example.io/...`.
    - Sometimes the Go language server's heuristics are not able to locate a canonical import path comment in a repository, in which case you can specify the root import path of your repository directly by placing a `.sourcegraph/config.json` file in the root of your repository, e.g.:

    ```json
    {
      "go": {
        "RootImportPath": "example.io/pkg"
      }
    }
    ```

    Which would tell the Go language server to clone `github.com/example/pkg` into `$GOPATH/src/example.io/pkg`.

3.  Otherwise, Sourcegraph will attempt to fetch `example.io/pkg/logger` via the network using `go get example.io/pkg/logger`.
