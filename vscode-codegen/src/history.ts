import * as vscode from "vscode";
import { getSymbols } from "./vscode-utils";
import { SymbolKind } from "vscode";
import { InflatedHistoryItem, HistoryItem, InflatedSymbol } from "@sourcegraph/cody-common";

export class History {
	window = 10;
	history: HistoryItem[];

	constructor() {
		this.history = [];
	}

	private addItem(item: HistoryItem) {
		if (item.uri.scheme === "codegen") {
			return;
		}
		if (this.history.length >= this.window) {
			this.history.shift();
		}
		this.history.push(item);
	}

	private lastItem(): HistoryItem | null {
		if (this.history.length === 0) {
			return null;
		}
		return this.history[this.history.length - 1];
	}

	async getInfo(): Promise<InflatedHistoryItem[]> {
		const contextItems = [];
		for (const item of this.history) {
			const doc = await vscode.workspace.openTextDocument(item.uri);
			const snippet = doc.getText(selectionAround(doc, item.selection));
			const symbols = await getSymbols(item.uri);
			const callableSymbols = symbols
				.filter(
					(s) =>
						s.kind in
						[
							SymbolKind.Class,
							SymbolKind.Function,
							SymbolKind.Method,
							SymbolKind.Interface,
							SymbolKind.Struct,
						]
				)
				.map((symbol) => inflateSymbol(doc, symbol));

			contextItems.push({
				item,
				snippet,
				symbols: callableSymbols,
			});
		}
		return contextItems;
	}

	register(context: vscode.ExtensionContext): void {
		context.subscriptions.push(
			vscode.window.onDidChangeActiveTextEditor((event) => {
				if (!event?.document.uri) {
					return;
				}
				this.addItem({
					uri: event.document.uri,
					selection: event.selection,
				});
			}),
			vscode.window.onDidChangeTextEditorSelection((event) => {
				if (!vscode.window.activeTextEditor?.document.uri) {
					return;
				}
				if (event.selections.length === 0) {
					return;
				}
				const item: HistoryItem = {
					uri: vscode.window.activeTextEditor?.document.uri,
					selection: event.selections[0],
				};
				const lastItem = this.lastItem();
				if (lastItem && isDupe(lastItem, item)) {
					return;
				}
				this.addItem(item);
			})
		);
	}
}

function isDupe(item1: HistoryItem, item2: HistoryItem) {
	return (
		item1.uri === item2.uri &&
		item1.selection.start.line &&
		isCloseOrOverlapping(item1.selection, item2.selection)
	);
}

const closeOrOverlappingThreshold = 3;
function isCloseOrOverlapping(
	range1: vscode.Range,
	range2: vscode.Range
): boolean {
	if (range1.intersection(range2)) {
		return true;
	}
	const lineDiff1 = Math.abs(range1.end.line - range2.start.line);
	const lineDiff2 = Math.abs(range2.end.line - range1.start.line);
	return (
		lineDiff1 < closeOrOverlappingThreshold ||
		lineDiff2 < closeOrOverlappingThreshold
	);
}

function selectionAround(
	document: vscode.TextDocument,
	range: vscode.Range
): vscode.Range {
	if (!range.start.isEqual(range.end)) {
		return range;
	}
	return document.validateRange(
		new vscode.Range(
			Math.max(0, range.start.line - 5),
			0,
			range.end.line + 5,
			0
		)
	);
}
export function inflateSymbol(
	doc: vscode.TextDocument,
	symbol: vscode.DocumentSymbol
): InflatedSymbol {
	return {
		symbol,
		text: doc.getText(symbol.range),
	};
}
