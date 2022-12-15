import * as vscode from "vscode";
import { TextDocumentContentProvider } from "vscode";

export class CodegenDocumentProvider implements TextDocumentContentProvider {
  loadedDocuments: { [name: string]: string } = {};

  fireDocumentChanged(uri: vscode.Uri): void {
    this.onDidChangeEmitter.fire(uri);
  }

  setDocument(uri: vscode.Uri, contents: Promise<string> | null): void {
    const uriStr = uri.toString();
    delete this.loadedDocuments[uriStr];
    if (!contents) {
      this.fireDocumentChanged(uri);
      return;
    }
    contents.then((loadedContents) => {
      this.loadedDocuments[uriStr] = loadedContents;
      this.fireDocumentChanged(uri);
    });
  }

  onDidChangeEmitter = new vscode.EventEmitter<vscode.Uri>();
  onDidChange = this.onDidChangeEmitter.event;

  provideTextDocumentContent(
    uri: vscode.Uri,
    token: vscode.CancellationToken
  ): string {
    if (!this.loadedDocuments[uri.toString()]) {
      return `// ${randomFrom(waitPhrases)}, ${randomFrom(
        waitPhraseSuffixes
      )}...`;
    }

    return this.loadedDocuments[uri.toString()];
  }
}

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

export function completionsDisplayString(
  completions:
    | { name: string; completions: vscode.InlineCompletionItem[] }[]
    | null
): string {
  if (!completions) {
    return `// ${randomFrom(waitPhrases)}, ${randomFrom(
      waitPhraseSuffixes
    )}...`;
  }
  const modelSeparator =
    "// ==============================================================";
  const completionSeparator =
    "// --------------------------------------------------------------";

  return (
    `/**\n * Suggestions:\n */\n\n` +
    completions
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
