# Running the sample language server via a Visual Studio code extension

This describes how to run the sample language server in `../sample`
in Visual Studio Code via the Visual Studio Code extension in this directory.

To run the sample language server and extension:

1.  Run `npm install` in this directory (`vscode-client`).
1.  Open this directory by itself in Visual Studio Code.
1.  `cd ../sample`. Build the the `sample_server` program and
    ensure it is in the `$PATH` that Visual Studio Code sees
    (e.g., `go build -o /usr/local/bin/sample_server sample_server.go`).
1.  Follow the instructions in the
    [Node.js example language server tutorial](https://code.visualstudio.com/docs/extensions/example-language-server)
    under "To test the language server do the following:" to build and
    run the vscode-client sample extension.
1.  When the new Visual Studio Code window opens with the extension,
    create a new text file and type some text. Move your mouse cursor
    over the text, and you should see "Hello over LSP!" in a tooltip.
    Right click on any text and click "Go to Definition", and you'll
    be taken to the beginning of the file. Right click on any text and
    click "Find All References", and if your text is longer than 6
    characters, you'll be shown two references in the first line of your
    file.

Debug tips if extension throws an error:

1.  Open the debug pane (Cmd+Shift+D) to see the call stack.
1.  Open the debug console (Cmd+Shift+Y) and press the green
    "play"/triangle button to see the error logs.
1.  Commonly seen error is the server is not in a `$PATH` that VSCode sees.

To run your own language server with this extension, see [`../sample/README.md`](../sample/README.md).
