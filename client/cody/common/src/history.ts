import * as vscode from 'vscode'
import { DocumentSymbol } from 'vscode'

export interface InflatedSymbol {
	symbol: DocumentSymbol
	text: string
}

export interface InflatedHistoryItem {
	item: HistoryItem
	snippet: string
	symbols: InflatedSymbol[]
}

export interface HistoryItem {
	uri: vscode.Uri
	selection: vscode.Selection
}
