import * as vscode from "vscode";
import { LLMDebugInfo } from "@sourcegraph/cody-common";

export class CompletionsDocumentProvider
	implements vscode.TextDocumentContentProvider, vscode.HoverProvider
{
	completionsByUri: { [uri: string]: CompletionGroup[] } = {};

	private fireDocumentChanged(uri: vscode.Uri): void {
		this.onDidChangeEmitter.fire(uri);
	}

	clearCompletions(uri: vscode.Uri) {
		delete this.completionsByUri[uri.toString()];
		this.fireDocumentChanged(uri);
	}

	addCompletions(
		uri: vscode.Uri,
		name: string,
		completions: string[],
		debug?: LLMDebugInfo
	) {
		if (!this.completionsByUri[uri.toString()]) {
			this.completionsByUri[uri.toString()] = [];
		}
		this.completionsByUri[uri.toString()].push({
			name,
			completions: completions.map((c) => ({ insertText: c })),
			debug,
		});
		this.fireDocumentChanged(uri);
	}

	onDidChangeEmitter = new vscode.EventEmitter<vscode.Uri>();
	onDidChange = this.onDidChangeEmitter.event;

	provideTextDocumentContent(uri: vscode.Uri): string {
		const completionGroups = this.completionsByUri[uri.toString()];
		if (!completionGroups) {
			return "// Loading...";
		}
		return completionGroups
			.map(({ name: modelName, completions }) =>
				completions
					.map((completion, i) => {
						const titleLine = `*** ${modelName} (${i + 1}/${
							completions.length
						}) ***`;
						return `/*${"*".repeat(Math.max(0, titleLine.length - 1))}
 ${titleLine}
 ${"*".repeat(Math.max(0, titleLine.length - 1))}*/
${completion.insertText}`;
					})
					.join(`\n\n`)
			)
			.join(`\n\n`);
	}

	provideHover(
		document: vscode.TextDocument,
		position: vscode.Position
	): vscode.ProviderResult<vscode.Hover> {
		const completionGroups = this.completionsByUri[document.uri.toString()];
		if (!completionGroups) {
			return null;
		}

		const wordRange = document.getWordRangeAtPosition(position, /[\w\-:]+/);
		if (!wordRange) {
			return null;
		}
		const word = document.getText(wordRange);
		for (const { name: modelName, debug } of completionGroups) {
			if (!debug) {
				continue;
			}

			if (modelName === word) {
				const rawPrompt = new vscode.MarkdownString(
					`<pre>${debug.prompt}</pre>`
				);
				rawPrompt.supportHtml = true;
				return new vscode.Hover([
					"Options:",
					new vscode.MarkdownString(
						`\`\`\`\n${JSON.stringify(debug.llmOptions, null, 2)}\n\`\`\``
					),
					"Prompt:",
					rawPrompt,
				]);
			}
		}
		return null;
	}
}

export interface CompletionGroup {
	name: string;
	completions: vscode.InlineCompletionItem[];
	debug?: LLMDebugInfo;
}
