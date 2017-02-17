import { BrowserLanguageClient } from "@sourcegraph/vscode-languageclient/lib/browser";

import { webSocketStreamOpener } from "sourcegraph/ext/lsp/connection";
import { getModes } from "sourcegraph/util/features";

export function activate(): void {
	// self.location is the blob: URI, so we need to get the main page location.
	let wsOrigin = self.location.origin.replace(/^https?:\/\//, (match) => {
		return match === "http://" ? "ws://" : "wss://";
	});

	getModes().forEach(mode => {
		const client = new BrowserLanguageClient("lsp", "lsp", webSocketStreamOpener(`${wsOrigin}/.api/lsp`), {
			documentSelector: [mode],
			initializationOptions: { mode },
		});
		client.start();
	});

}
