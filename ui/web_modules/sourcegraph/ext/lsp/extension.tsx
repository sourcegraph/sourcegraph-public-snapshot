import * as vscode from "vscode";

import { BrowserLanguageClient } from "@sourcegraph/vscode-languageclient/lib/browser";
import { v4 as uuidV4 } from "uuid";

import { isOnPremInstance } from "sourcegraph/app/context";
import { MessageTrace, webSocketStreamOpener } from "sourcegraph/ext/lsp/connection";
import { InitializationOptions } from "sourcegraph/ext/protocol";
import { Features, getModes } from "sourcegraph/util/features";
import { inventoryLangToMode } from "sourcegraph/util/inventory";

export function activate(): void {
	// self.location is the blob: URI, so we need to get the main page location.
	let wsOrigin = self.location.origin.replace(/^https?:\/\//, (match) => {
		return match === "http://" ? "ws://" : "wss://";
	});

	const initOpts: InitializationOptions = (self as any).extensionHostOptions;
	const langs = new Set<string>((initOpts.langs || []).map(inventoryLangToMode));

	getModes(isOnPremInstance(initOpts.context.authEnabled)).forEach(mode => {
		if (!langs.has(mode)) {
			return;
		}

		const workspaceUri = vscode.Uri.parse(initOpts.workspace);

		// We include ?mode= in the url to make it easier to find the correct LSP websocket connection.
		// It does not affect any behaviour.
		const client = new BrowserLanguageClient("lsp-" + mode, "lsp-" + mode, webSocketStreamOpener(`${wsOrigin}/.api/lsp?mode=${mode}`, getRequestTracer(mode)), {
			documentSelector: [mode],
			initializationOptions: {
				mode,
				rev: initOpts.revState!.commitID,
				session: generateLSPSessionKeyIfNeeded(),
			},
			uriConverters: {
				code2Protocol: (value: vscode.Uri) => {
					if (value.scheme === "file") {
						let filePath = value.toString().substr(initOpts.workspace.length + 1); // trim leading "/" after workspace path; possibly empty
						// TODO(john): if workspace rev state changes, we re-open a LSP connection with the new revision base.
						return workspaceUri.with({ scheme: "git", query: initOpts.revState!.commitID, fragment: filePath }).toString();
					}
					return value.toString();
				},
				protocol2Code: (value: string) => {
					const uri = vscode.Uri.parse(value);
					if (uri.scheme === "git") {
						// convert to file if in the same workspace
						if (uri.with({ scheme: "file", query: "", fragment: "" }).toString() === initOpts.workspace) {
							return uri.with({ scheme: "file", query: "", path: uri.path + `${uri.fragment !== "" ? `/${uri.fragment}` : ""}`, fragment: "" });
						}
					}
					return vscode.Uri.parse(value);
				},
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
	const initOpts: InitializationOptions = (self as any).extensionHostOptions;
	if (initOpts.revState && initOpts.revState.zapRef) {
		return uuidV4(); // uses a cryptographic RNG on browsers with window.crypto (all modern browsers)
	}
	return undefined;
}

// tslint:disable: no-console

function getRequestTracer(mode: string): ((trace: MessageTrace) => void) | undefined {
	if (!Features.trace.isEnabled() || !(global as any).console) {
		return undefined;
	}
	const console = (global as any).console;
	if (!console.debug || !console.group) {
		return undefined;
	}
	return (trace: MessageTrace) => {
		console.groupCollapsed(
			"%c%s%c LSP %s %s %c%sms",
			`background-color:${trace.response.error ? "red" : "green"};color:white`, trace.response.error ? "ERR" : "OK",
			`background-color:inherit;color:inherit;`,
			mode,
			describeRequest(trace.request.method, trace.request.params),
			"color:#999;font-weight:normal;font-style:italic",
			trace.endTime - trace.startTime,
		);
		if (trace.response.meta && trace.response.meta["X-Trace"]) {
			console.debug("Trace: %s", trace.response.meta["X-Trace"]);
		}
		console.debug("Request Params: %O", trace.request.params);
		console.debug("Response: %O", trace.response);
		console.groupEnd();
	};
}

function describeRequest(method: string, params: any): string {
	if (params.textDocument && params.textDocument.uri && params.position) {
		return `${method} @ ${params.position.line + 1}:${params.position.character + 1}`;
	}
	if (typeof params.query !== "undefined") {
		return `${method} with query ${JSON.stringify(params.query)}`;
	}
	return method;
}
