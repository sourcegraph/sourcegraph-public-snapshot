import * as vscode from "vscode";
import {
  TextDocument,
  Position,
  InlineCompletionContext,
  CancellationToken,
  ProviderResult,
  InlineCompletionItem,
  InlineCompletionList,
  DocumentSymbol,
  SymbolInformation,
  SymbolKind,
} from "vscode";
import * as openai from "openai";
import { cp } from "fs";
import { CompletionSupplier } from "./models/model";
import { OpenAICompletionSupplier } from "./models/codegen";
import {
  CodegenDocumentProvider,
  completionsDisplayString,
} from "./docprovider";

export async function generateTestFromSelection(
  documentProvider: CodegenDocumentProvider
) {
  const editor = vscode.window.activeTextEditor;
  const ext = editor?.document.fileName.split(".").pop();
  const snippet = editor?.document.getText(editor.selection);
  const prompt = `Here is the code I wish to test:
  \`\`\`
  ${snippet}
  \`\`\`

  Write the code for the unit test:
  \`\`\`
  `;

  const fileUri = vscode.window.activeTextEditor?.document.uri;
  if (fileUri) {
    getSymbols(fileUri);
  }

  const completions = getCompletions(prompt);

  const generatedUri = vscode.Uri.parse(`codegen:unittests.${ext}`);
  documentProvider.setDocument(generatedUri, completions);

  const doc = await vscode.workspace.openTextDocument(generatedUri);
  await vscode.window.showTextDocument(doc, {
    preview: false,
    viewColumn: 2,
  });
}

function getCompletions(prompt: string) {
  const config = new openai.Configuration({
    apiKey: vscode.workspace
      .getConfiguration()
      .get("conf.codebot.openai.apiKey"),
  });
  const oa = new openai.OpenAIApi(config);
  const completions = oa
    .createCompletion({
      model: "code-davinci-002",
      prompt: prompt,
      temperature: 0.2,
      max_tokens: 700,
      stop: "```",
      n: 1,
    })
    .then((response) => response.data.choices.map((choice) => choice.text))
    .then((choiceText) => choiceText.filter((c) => c).join("\n\n"));
  return completions;
}

export async function generateTest(documentProvider: CodegenDocumentProvider) {
  const editor = vscode.window.activeTextEditor;
  const ext = editor?.document.fileName.split(".").pop();
  if (!editor?.document.uri) {
    return;
  }
  const symbols = await getSymbols(editor.document.uri);

  symbols.sort((a, b) => {
    let diff = 0;
    diff += 4 * ((isTestable(a) ? 0 : 1) - (isTestable(b) ? 0 : 1));
    diff += 2 * ((isProbablyTest(a) ? 0 : 1) - (isProbablyTest(b) ? 0 : 1));
    diff += 1 * ((isClose(a) ? 0 : 1) - (isClose(b) ? 0 : 1));
    return diff;
  });

  const qp = await vscode.window.createQuickPick();
  qp.show();
  const selected = await new Promise(async (resolve) => {
    qp.title = "select the function to test";
    qp.items = symbols.map((s) => ({ label: s.name }));
    qp.onDidChangeSelection((s) => resolve(s.length > 0 && s[0].label));
    qp.onDidChangeActive(() => {
      if (qp.activeItems.length === 0) {
        return;
      }

      const symbolName = qp.activeItems[0].label;
      const range = symbols.find((s) => s.name === symbolName)?.selectionRange;
      if (range) {
        editor.revealRange(range, vscode.TextEditorRevealType.AtTop);
      }
    });
  });

  if (!selected) {
    return; // aborted
  }

  const selectedSymbol = symbols.find((s) => s.name === selected);
  const symbolCode = editor.document.getText(selectedSymbol?.range);

  const prompt = `\`\`\`
  ${symbolCode}
  \`\`\`
  Write the unit test for the above code:
  \`\`\
  `;

  const completions = getCompletions(prompt);

  const generatedUri = vscode.Uri.parse(`codegen:unittests.${ext}`);
  documentProvider.setDocument(generatedUri, completions);

  const doc = await vscode.workspace.openTextDocument(generatedUri);
  await vscode.window.showTextDocument(doc, {
    preview: false,
    viewColumn: 2,
  });
}

async function getSymbols(uri: vscode.Uri): Promise<DocumentSymbol[]> {
  const symbols: (SymbolInformation | DocumentSymbol)[] =
    await vscode.commands.executeCommand(
      "vscode.executeDocumentSymbolProvider",
      uri
    );

  if (!symbols) {
    throw new Error(
      "No symbols found. Install an editor extension that supplies symbols"
    );
  }

  const docSymbols: DocumentSymbol[] = [];
  for (const s of symbols) {
    const ds = s as DocumentSymbol;
    if (!ds.range) {
      throw new Error(
        `Found SymbolInformation ${s.name}, expects only DocumentSymbol instances`
      );
    }
    docSymbols.push(ds);
  }
  return flattenSymbols(docSymbols);
}

function flattenSymbols(symbols: DocumentSymbol[]): DocumentSymbol[] {
  return symbols.flatMap((s) => [s, ...s.children]);
}

function isProbablyTest(s: DocumentSymbol): boolean {
  if (s.name.toLowerCase().indexOf("test") !== -1) {
    return true;
  }
  if (s.name === "describe") {
    return true;
  }
  return false;
}

function isTestable(s: DocumentSymbol): boolean {
  return [SymbolKind.Function, SymbolKind.Method].indexOf(s.kind) !== -1;
}

function isClose(s: DocumentSymbol): boolean {
  const curPoint = vscode.window.activeTextEditor?.selection.end;
  if (!curPoint) return false;
  return s.range.contains(curPoint);
}
