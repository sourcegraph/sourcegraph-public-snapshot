import * as vscode from 'vscode'
import { DocumentSymbol, SymbolInformation } from 'vscode'

export async function getSymbols(uri: vscode.Uri): Promise<DocumentSymbol[]> {
	const symbols: (SymbolInformation | DocumentSymbol)[] = await vscode.commands.executeCommand(
		'vscode.executeDocumentSymbolProvider',
		uri
	)

	if (!symbols) {
		throw new Error(`No symbols found for ${uri.toString()}. Install an editor extension that supplies symbols`)
	}

	const docSymbols: DocumentSymbol[] = []
	for (const symbol of symbols) {
		const docSymbol = symbol as DocumentSymbol
		if (!docSymbol.range) {
			throw new Error(`Found SymbolInformation ${symbol.name}, expects only DocumentSymbol instances`)
		}
		docSymbols.push(docSymbol)
	}
	return flattenSymbols(docSymbols)
}

function flattenSymbols(symbols: DocumentSymbol[]): DocumentSymbol[] {
	return symbols.flatMap(symbol => [symbol, ...symbol.children])
}
