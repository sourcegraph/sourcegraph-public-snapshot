import { LanguageClient } from "vscode-languageclient/lib/client";

import { webSocketStreamOpener } from "sourcegraph/ext/lsp/connection";

export function activate(): void {
	// self.location is the blob: URI, so we need to get the main page location.
	let wsOrigin = self.location.origin.replace(/^https?:\/\//, (match) => {
		return match === "http://" ? "ws://" : "wss://";
	});
	const client = new LanguageClient("lsp", "lsp", webSocketStreamOpener(`${wsOrigin}/.api/lsp`), {
		documentSelector: ["go"],
		initializationOptions: { mode: "go" },
	});
	client.start();
}
