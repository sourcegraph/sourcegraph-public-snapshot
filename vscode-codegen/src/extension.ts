import * as vscode from "vscode";
import {
  TextDocument,
  Position,
  InlineCompletionContext,
  CancellationToken,
  ProviderResult,
  InlineCompletionItem,
  InlineCompletionList,
} from "vscode";
import * as openai from "openai";
import { cp } from "fs";
import { CompletionSupplier } from "./models/model";
import { CodeGenCompletionSupplier } from "./models/codegen";

const log = (...args: any[]) => console.log(...args);
const waitPhrases = [
  "Spelunking through latent space",
  "Reticulating neural splines",
  "Conferring with the robots",
  "Rummaging through tensors",
  "Rousting the neural nets",
  "Munging the perceptrons",
  "Rectifying the sigmoids",
  "Monkeying around with bits",
  "Bitlifying your monkey language",
];
const waitPhraseSuffixes = [
  "wait a sec",
  "just a moment",
  "hold tight",
  "almost ready",
  "thank you for your patience",
];
const randomFrom = (arr: string[]): string => {
  return arr[Math.floor(Math.random() * arr.length)];
};

// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {
  const codegenCompletionProvider = new CompletionProvider([
    new CodeGenCompletionSupplier(
      "http://localhost:5000/v1",
      "fastertransformer", // alt would be Python-based (consult fauxpilot docs)
      log
    ),
  ]);
  codegenCompletionProvider.register(context);
}

// This method is called when your extension is deactivated
export function deactivate() {}

class CompletionProvider
  implements
    vscode.InlineCompletionItemProvider,
    vscode.TextDocumentContentProvider
{
  completionSuppliers: CompletionSupplier[];
  constructor(completionSuppliers: CompletionSupplier[]) {
    this.completionSuppliers = completionSuppliers;
  }

  register(context: vscode.ExtensionContext) {
    context.subscriptions.push(
      vscode.languages.registerInlineCompletionItemProvider(
        { pattern: "**" },
        this
      )
    );
    context.subscriptions.push(
      vscode.workspace.registerTextDocumentContentProvider("codegen", this)
    );

    context.subscriptions.push(
      vscode.commands.registerCommand("vscode-codegen.ai-suggest", () =>
        this.executeSuggestCommand()
      )
    );
  }

  pendingManualCompletions: Promise<
    {
      name: string;
      completions: InlineCompletionItem[];
    }[]
  > | null = null;
  async executeSuggestCommand(): Promise<void> {
    const currentEditor = vscode.window.activeTextEditor;
    if (!currentEditor) {
      return;
    }
    if (currentEditor.document.uri.scheme === "codegen") {
      return;
    }

    const filename = currentEditor.document.fileName;
    const ext = filename.slice(filename.lastIndexOf(".") + 1);
    const completionsUri = vscode.Uri.parse(`codegen:completions.${ext}`);
    this.pendingManualCompletions = null;
    this.onDidChangeEmitter.fire(completionsUri);

    await vscode.workspace
      .openTextDocument(completionsUri)
      .then((doc) =>
        vscode.window.showTextDocument(doc, { preview: false, viewColumn: 2 })
      );

    const position = currentEditor.selection.active;

    const theseCompletionsPromise = Promise.all(
      this.completionSuppliers.map(async (supplier) => {
        const completions = await supplier.getCompletions(
          currentEditor.document,
          position,
          256,
          5
        );
        return {
          name: supplier.getName(),
          completions,
        };
      })
    );
    this.pendingManualCompletions = theseCompletionsPromise;
    theseCompletionsPromise.then(() => {
      this.onDidChangeEmitter.fire(completionsUri);
    });
  }

  // Update the virtual document when new completions are available
  onDidChangeEmitter = new vscode.EventEmitter<vscode.Uri>();
  onDidChange = this.onDidChangeEmitter.event;
  async provideTextDocumentContent(uri: vscode.Uri): Promise<string> {
    if (!this.pendingManualCompletions) {
      return `// ${randomFrom(waitPhrases)}, ${randomFrom(
        waitPhraseSuffixes
      )}...`;
    }

    const completionsByModel = await this.pendingManualCompletions;
    const modelSeparator =
      "// ==============================================================";
    const completionSeparator =
      "// --------------------------------------------------------------";
    return (
      `/**\n * Suggestions:\n */\n\n` +
      completionsByModel
        .map(
          ({ name: modelName, completions }) =>
            `// Model ${modelName}\n\n` +
            completions
              .map((completion) => completion.insertText)
              .join(`\n\n${completionSeparator}\n\n`)
        )
        .join(`\n\n${modelSeparator}\n\n`)
    );
  }

  lastAutoSuggestRequestTime: number = 0;
  async provideInlineCompletionItems(
    document: TextDocument,
    position: Position,
    context: InlineCompletionContext,
    token: CancellationToken
  ): Promise<InlineCompletionItem[]> {
    // debounce
    const requestTime = Date.now();
    this.lastAutoSuggestRequestTime = requestTime;
    await new Promise((resolve) => setTimeout(resolve, 1000));
    if (requestTime !== this.lastAutoSuggestRequestTime) {
      return [];
    }

    return Promise.all(
      this.completionSuppliers.map((supplier) =>
        supplier.getCompletions(document, position, 128, 1)
      )
    ).then((allCompletions) => allCompletions.flatMap((c) => c));
  }
}
