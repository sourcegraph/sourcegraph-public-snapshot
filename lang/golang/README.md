# Running this Go language server with vscode

1. Build `langserver-go` and install it into your `$PATH`: `go install ./cmd/langserver-go`.
1. Follow the [instructions to install and run the vscode-client extension](../vscode-client/README.md).
1. Open a .go file and hover over an identifier. You should see a tooltip and documentation. The documentation currently is duplicated (both plaintext and HTML); this is a known issue.

To view langserver-go's stderr in vscode, select View → Output.
