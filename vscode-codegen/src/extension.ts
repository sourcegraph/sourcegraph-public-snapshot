import * as vscode from "vscode";
import { CompletionsDocumentProvider } from "./docprovider";
import { History } from "./history";
import { ChatViewProvider, WSChatClient } from "./chat";
import { WSCompletionsClient, fetchAndShowCompletions } from "./completions";

// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
export async function activate(context: vscode.ExtensionContext) {
	console.log("codebot extension activated");
	const settings = vscode.workspace.getConfiguration();
	const documentProvider = new CompletionsDocumentProvider();
	const history = new History();
	history.register(context);

	const wsCompletionsClient = await WSCompletionsClient.new(
		"ws://localhost:8080/completions"
	);
	const wsChatClient = await WSChatClient.new("ws://localhost:8080/chat");

	context.subscriptions.push(
		vscode.workspace.registerTextDocumentContentProvider(
			"codegen",
			documentProvider
		),
		vscode.languages.registerHoverProvider(
			{ scheme: "codegen" },
			documentProvider
		),

		vscode.commands.registerCommand("vscode-codegen.ai-suggest", async () => {
			await fetchAndShowCompletions(
				wsCompletionsClient,
				documentProvider,
				history
			);
		}),
		// vscode.commands.registerCommand(
		// 	"codebot.generate-test-from-selection",
		// 	() => generateTestFromSelection(documentProvider)
		// ),
		// vscode.commands.registerCommand("codebot.generate-test", () =>
		// 	generateTest(documentProvider)
		// ),

		vscode.window.registerWebviewViewProvider(
			"cody.chat",
			new ChatViewProvider(wsChatClient)
		)
	);
}

// This method is called when your extension is deactivated
export function deactivate() {}
