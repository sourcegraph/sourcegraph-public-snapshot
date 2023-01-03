import * as vscode from "vscode";
import { DocumentSymbol, SymbolKind } from "vscode";

export interface InflatedSymbol {
	symbol: DocumentSymbol;
	text: string;
}

export interface InflatedHistoryItem {
	item: HistoryItem;
	snippet: string;
	symbols: InflatedSymbol[];
}

interface HistoryItem {
	uri: vscode.Uri;
	selection: vscode.Selection;
}
