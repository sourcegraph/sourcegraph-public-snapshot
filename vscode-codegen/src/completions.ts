import * as vscode from "vscode";
import { CompletionsDocumentProvider } from "./docprovider";
import { History } from "./history";
import { getReferences } from "./autocomplete/completion-provider";
import {
	CompletionsArgs,
	LLMDebugInfo,
	WSCompletionResponse,
	WSCompletionsRequest,
} from "@sourcegraph/cody-common";
import { WSClient } from "./wsclient";

interface CompletionCallbacks {
	onCompletions(completions: string[], debugInfo?: LLMDebugInfo): void;
	onMetadata(metadata: any): void;
	onDone(): void;
	onError(err: string): void;
}

export class WSCompletionsClient {
	static async new(addr: string): Promise<WSCompletionsClient> {
		const wsclient = await WSClient.new<
			Omit<WSCompletionsRequest, "requestId">,
			WSCompletionResponse
		>(addr);
		return new WSCompletionsClient(wsclient);
	}

	private constructor(
		private wsclient: WSClient<
			Omit<WSCompletionsRequest, "requestId">,
			WSCompletionResponse
		>
	) {}

	async getCompletions(
		args: CompletionsArgs,
		callbacks: CompletionCallbacks
	): Promise<void> {
		await this.wsclient.sendRequest(
			{
				kind: "getCompletions",
				args,
			},
			(resp) => {
				try {
					switch (resp.kind) {
						case "completion":
							callbacks.onCompletions(resp.completions, resp.debugInfo);
							return false;
						case "metadata":
							callbacks.onMetadata(resp.metadata);
							return false;
						case "error":
							callbacks.onError(resp.error);
							return false;
						case "done":
							callbacks.onDone();
							return true;
						default:
							return false;
					}
				} catch (error: any) {
					vscode.window.showErrorMessage(error);
					return false;
				}
			}
		);
	}
}

async function getCompletionsArgs(history: History): Promise<CompletionsArgs> {
	const currentEditor = vscode.window.activeTextEditor;
	if (!currentEditor) {
		throw new Error("no current active editor");
	}

	const document = currentEditor.document;

	const position = currentEditor.selection.active;
	const prefixRange = new vscode.Range(0, 0, position.line, position.character);

	const prefix = document.getText(prefixRange);
	const historyInfo = await history.getInfo();
	const references = await getReferences(document, position, [
		new vscode.Location(document.uri, prefixRange),
	]);

	return {
		history: historyInfo,
		prefix,
		references,
		uri: document.uri.toString(),
	};
}

export async function fetchAndShowCompletions(
	wsclient: WSCompletionsClient,
	documentProvider: CompletionsDocumentProvider,
	history: History
) {
	const currentEditor = vscode.window.activeTextEditor;
	if (!currentEditor || currentEditor?.document.uri.scheme === "codegen") {
		return;
	}
	const filename = currentEditor.document.fileName;
	const ext = filename.split(".").pop();
	const completionsUri = vscode.Uri.parse(`codegen:completions.${ext}`);
	documentProvider.clearCompletions(completionsUri);
	vscode.workspace.openTextDocument(completionsUri).then((doc) => {
		vscode.window.showTextDocument(doc, {
			preview: false,
			viewColumn: 2,
		});
	});

	try {
		await wsclient.getCompletions(await getCompletionsArgs(history), {
			onCompletions: function (
				completions: string[],
				debug?: LLMDebugInfo | undefined
			): void {
				const name = "openai-completions";
				documentProvider.addCompletions(
					completionsUri,
					name,
					completions,
					debug
				);
			},
			onMetadata: function (metadata: any): void {
				console.log(`received metadata ${metadata}`);
			},
			onDone: function (): void {
				console.log("received done");
			},
			onError: function (err: string): void {
				console.error(`received error ${err}`);
			},
		});
	} catch (error: any) {
		vscode.window.showErrorMessage(error);
	}
}
