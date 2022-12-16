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
import { OpenAICompletionSupplier } from "./models/codegen";
import { generateTestFromSelection, generateTest } from "./testgen";
import {
  CodegenDocumentProvider,
  completionsDisplayString,
} from "./docprovider";

const log = (...args: any[]) => console.log(...args);

// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {
  const documentProvider = new CodegenDocumentProvider();

  const codegenCompletionProvider = new CompletionProvider(
    [
      // new OpenAICompletionSupplier(
      //   new openai.Configuration({
      //     basePath: "http://localhost:5000/v1",
      //   }),
      //   "fastertransformer", // alt would be Python-based (consult fauxpilot docs)
      //   "CodeGen",
      //   log
      // ),
      new OpenAICompletionSupplier(
        new openai.Configuration({
          apiKey: vscode.workspace
            .getConfiguration()
            .get("conf.codebot.openai.apiKey"),
        }),
        "code-davinci-002",
        "Codex (code-davinci-002)",
        log
      ),
    ],
    documentProvider
  );

  context.subscriptions.push(
    vscode.workspace.registerTextDocumentContentProvider(
      "codegen",
      documentProvider
    ),
    vscode.languages.registerInlineCompletionItemProvider(
      { pattern: "**" },
      codegenCompletionProvider
    ),
    vscode.commands.registerCommand("vscode-codegen.ai-suggest", () =>
      codegenCompletionProvider.executeSuggestCommand()
    ),
    vscode.commands.registerCommand(
      "codebot.generate-test-from-selection",
      () => generateTestFromSelection(documentProvider)
    ),
    vscode.commands.registerCommand("codebot.generate-test", () =>
      generateTest(documentProvider)
    )
  );
}

// This method is called when your extension is deactivated
export function deactivate() {}

class CompletionProvider implements vscode.InlineCompletionItemProvider {
  completionSuppliers: CompletionSupplier[];
  documentProvider: CodegenDocumentProvider;
  constructor(
    completionSuppliers: CompletionSupplier[],
    documentProvider: CodegenDocumentProvider
  ) {
    this.completionSuppliers = completionSuppliers;
    this.documentProvider = documentProvider;
  }

  async executeSuggestCommand(): Promise<void> {
    const currentEditor = vscode.window.activeTextEditor;
    if (!currentEditor) {
      return;
    }
    if (currentEditor.document.uri.scheme === "codegen") {
      return;
    }

    const filename = currentEditor.document.fileName;
    const ext = filename.split(".").pop();
    const completionsUri = vscode.Uri.parse(`codegen:completions.${ext}`);

    const docOpener = vscode.workspace
      .openTextDocument(completionsUri)
      .then((doc) => {
        vscode.window.showTextDocument(doc, {
          preview: false,
          viewColumn: 2,
        });
      });

    const position = currentEditor.selection.active;
    const completionsPromise = Promise.all(
      this.completionSuppliers.map(async (supplier) => {
        const completions = await supplier.getCompletions(
          currentEditor.document,
          position,
          150, // 256
          3
        );
        return {
          name: supplier.getName(),
          completions,
        };
      })
    ).then((completions) => completionsDisplayString(completions));

    this.documentProvider.setDocument(completionsUri, completionsPromise);

    await docOpener;
  }

  // Update the virtual document when new completions are available
  onDidChangeEmitter = new vscode.EventEmitter<vscode.Uri>();
  onDidChange = this.onDidChangeEmitter.event;

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
