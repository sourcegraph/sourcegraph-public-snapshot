// sc// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';
import { TextDocument, Position, InlineCompletionContext, CancellationToken, ProviderResult, InlineCompletionItem, InlineCompletionList } from 'vscode';
import * as openai from 'openai';

const serverUrl = "http://localhost:5000/v1";
const model = "fastertransformer";

const log = (...args: any[]) => console.log(...args)
const waitPhrases = [
	"Spelunking through latent space",
	"Reticulating neural splines",
	"Conferring with the robots",
	"Rummaging through tensors",
	"Rousting the neural nets",
	"Munging the perceptrons",
	"Rectifying the sigmoids",
	"Monkeying around with bits",
	"Bitlifying your monkey language"
]
const waitPhraseSuffixes = [
	"wait a sec",
	"just a moment",
	"hold tight",
	"almost ready",
	"thank you for your patience",
]
const randomFrom = (arr: string[]): string => {
	return arr[Math.floor(Math.random() * arr.length)]
}

// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {
	const codegenCompletionProvider = new CodegenCompletionProvider()
	codegenCompletionProvider.register(context)
}


// This method is called when your extension is deactivated
export function deactivate() {}


class CodegenCompletionProvider implements vscode.InlineCompletionItemProvider, vscode.TextDocumentContentProvider {

	register(context: vscode.ExtensionContext) {
		context.subscriptions.push(vscode.languages.registerInlineCompletionItemProvider(
			{pattern: '**'},
			this,
		))
		context.subscriptions.push(vscode.workspace.registerTextDocumentContentProvider("codegen", this))

		context.subscriptions.push(vscode.commands.registerCommand("vscode-codegen.ai-suggest", () => this.executeSuggestCommand()))
	}

    pendingManualCompletions: Promise<InlineCompletionItem[]> | null = null
	async executeSuggestCommand(): Promise<void> {
		const currentEditor = vscode.window.activeTextEditor
		if (!currentEditor) {
			return
		}
		if (currentEditor.document.uri.scheme === "codegen") {
			return
		}

		const filename = currentEditor.document.fileName
		const ext = filename.slice(filename.lastIndexOf(".")+1)
		const completionsUri = vscode.Uri.parse(`codegen:completions.${ext}`)
		this.pendingManualCompletions = null
		this.onDidChangeEmitter.fire(completionsUri)

		await vscode.workspace.openTextDocument(completionsUri).then(doc => vscode.window.showTextDocument(doc, { preview: false, viewColumn: 2 }))

		const position = currentEditor.selection.active
		const theseCompletions = this.getCompletions(currentEditor.document, position, 256, 5)
		this.pendingManualCompletions = theseCompletions
		theseCompletions.then(() => {
			this.onDidChangeEmitter.fire(completionsUri)
		})
	}

	onDidChangeEmitter = new vscode.EventEmitter<vscode.Uri>()
	onDidChange = this.onDidChangeEmitter.event
	async provideTextDocumentContent(uri: vscode.Uri): Promise<string> {
		if (!this.pendingManualCompletions) {
			return `// ${randomFrom(waitPhrases)}, ${randomFrom(waitPhraseSuffixes)}...`
		}

		const completions = await this.pendingManualCompletions
		const separator = "// =============================================================="
		return `/**\n * Suggestions:\n */\n\n` + completions.map(completion => completion.insertText).join(`\n\n${separator}\n\n`)
	}

	lastAutoSuggestRequestTime: number = 0;
	async provideInlineCompletionItems(document: TextDocument, position: Position, context: InlineCompletionContext, token: CancellationToken): Promise<InlineCompletionItem[]> {
		// debounce
		const requestTime = Date.now();
		this.lastAutoSuggestRequestTime = requestTime;
		await new Promise(resolve => setTimeout(resolve, 1000));
		if (requestTime !== this.lastAutoSuggestRequestTime) {
			return []
		}
		return this.getCompletions(document, position, 128, 1)
	}

	async getCompletions(document: TextDocument, position: Position, maxTokens: number, tries: number): Promise<InlineCompletionItem[]> {
		log("getCompletions:", document.uri, position, tries)
		const prompt = document.getText(new vscode.Range(0, 0, position.line, position.character));
		const oa = new openai.OpenAIApi(new openai.Configuration({ basePath: serverUrl }));
		const args = {
			model: model,
			prompt: prompt,
			max_tokens: maxTokens,
			stop: getStopSequences(position.line, document),
			n: tries,
		};
		const response = await oa.createCompletion(args);
		const completions = response.data.choices?.map(choice => choice.text).map(
		  choiceText =>
			new vscode.InlineCompletionItem(
			  choiceText as string,
			  new vscode.Range(position, position),
			),
		) || []
	
		log(`got ${completions.length} completions`)
		return completions
	}
	
}

export function getStopSequences(position: number, doc: vscode.TextDocument) {
  const stopSequences: string[] = [];

  // If the position is the last line, we don't use stop sequences
  if (doc.lineCount === position + 1) {
    return stopSequences;
  }
  // If the nextLine is not empty, it will be used as a stop sequence
  const nextLine = doc.lineAt(position + 1).text;
  if (nextLine) {
    stopSequences.push(nextLine.toString());
  }
//   log.debug(`Stop sequences: ${stopSequences.toString()}`);
  return stopSequences;
}
