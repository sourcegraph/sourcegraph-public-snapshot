import { v4 as uuidV4 } from "uuid";

import { BrowserLanguageClient } from "@sourcegraph/vscode-languageclient/lib/browser";

import { webSocketStreamOpener } from "sourcegraph/ext/lsp/connection";
import { Features, getModes } from "sourcegraph/util/features";

export function activate(): void {
	// self.location is the blob: URI, so we need to get the main page location.
	let wsOrigin = self.location.origin.replace(/^https?:\/\//, (match) => {
		return match === "http://" ? "ws://" : "wss://";
	});

	getModes().forEach(mode => {
		// We include ?mode= in the url to make it easier to find the correct LSP websocket connection.
		// It does not affect any behaviour.
		const client = new BrowserLanguageClient("lsp", "lsp", webSocketStreamOpener(`${wsOrigin}/.api/lsp?mode=${mode}`), {
			documentSelector: [mode],
			initializationOptions: {
				mode,
				session: generateLSPSessionKeyIfNeeded(),
			},
		});
		client.start();
	});

}

/**
 * generateLSPSessionIfNeeded generates a unique, difficult-to-guess
 * session key to pass to the LSP client proxy, which isolates this
 * session from others that target the same (mode, rootURI).
 *
 * Sessions with a nonempty session key can perform text edits (e.g.,
 * textDocument/didChange), so we must set it if document text will
 * ever need to change from what's in the Git commit specified in the
 * workspace root URI. Currently this is only needed for when Zap is
 * enabled.
 *
 * 🚨🚨 SECURITY: Anyone with this session key (and access to the
 * repository that this LSP session is for) can read the contents of
 * all files in the workspace and modify file contents in a way that
 * could lead to the user committing those modified contents (e.g., if
 * the user was in a live Zap session and then ran `git commit -a`
 * locally).
 */
function generateLSPSessionKeyIfNeeded(): string | undefined {
	if (!Features.zap.isEnabled()) {
		return undefined;
	}
	return uuidV4(); // uses a cryptographic RNG on browsers with window.crypto (all modern browsers)
}
