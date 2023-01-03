import * as vscode from "vscode";
import { DocumentSymbol, SymbolInformation } from "vscode";

export async function getSymbols(uri: vscode.Uri): Promise<DocumentSymbol[]> {
  const symbols: (SymbolInformation | DocumentSymbol)[] =
    await vscode.commands.executeCommand(
      "vscode.executeDocumentSymbolProvider",
      uri
    );

  if (!symbols) {
    throw new Error(
      `No symbols found for ${uri.toString()}. Install an editor extension that supplies symbols`
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
