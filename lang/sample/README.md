# Running your own language server

This describes how to run your own language server with the Visual Studio Code
extension in `../vscode-client`.  For instructions on running the sample
language server, see
[`../vscode-client/README.md`](../vscode-client/README.md).

To run your own language server:

1.  In the extension code (`../vscode-client/src/extension.ts`), in the
    `activate` function, change the `run` and `debug` field commands from
    `sample_server` to `${your_server}`.
1.  Open ../lang/vscode-client/package.json and under `"activationEvents"` add
    `"onCommand:${your_language}"`
1.  Build your server program in the `$PATH` that visual studio sees (e.g.
    `go build -o /usr/local/bin/${your_server} ${your_server}.go`.)
1.  Follow the instructions in the
    [Node.js example language server tutorial](https://code.visualstudio.com/docs/extensions/example-language-server)
    under "To test the language server do the following:" to build and run the
    vscode-client sample extension.
