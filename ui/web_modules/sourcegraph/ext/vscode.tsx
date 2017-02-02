import { createExtensionAPI } from "sourcegraph/ext/api";
import isWebWorker from "sourcegraph/util/isWebWorker";

// This file's exported API must be kept in sync with that of
// vscode.d.ts.
//
// We need to explicitly export every identifier here because
// extensions use `import * as vscode from "vscode";` or `import {Foo}
// from "vscode";` to import parts of the vscode extension API. In
// order for those import statements to work, they need to be
// individually named, top-level exports. There are 2 other possible
// solutions that seem nicer but don't work: (1) TypeScript `export =
// myObject;` (which is incompatible with the ES6 module target) and
// (2) adding exports to `__webpack_exports__` at runtime (which is
// not typesafe and produces unsuppressable webpack console warnings).

if (!isWebWorker) {
	throw new Error("vscode extension API must not be imported in main process (only extension host Web Worker)");
}

self["require"] = () => { throw new Error("require is not provided"); };

const vscode = createExtensionAPI()({ name: "global", id: "global", version: "0.0.0" } as any);

// Attach to global scope for easy/fun JavaScript console hacking.
self["vscode"] = vscode;

export const version = vscode.version;
// namespaces
export const commands = vscode.commands;
export const env = vscode.env;
export const extensions = vscode.extensions;
export const window = vscode.window;
export const languages = vscode.languages;
export const workspace = vscode.workspace;
// constructors
export const CancellationTokenSource = vscode.CancellationTokenSource;
export const CodeLens = vscode.CodeLens;
export const CompletionItem = vscode.CompletionItem;
export const CompletionItemKind = vscode.CompletionItemKind;
export const CompletionList = vscode.CompletionList;
export const Diagnostic = vscode.Diagnostic;
export const DiagnosticSeverity = vscode.DiagnosticSeverity;
export const Disposable = vscode.Disposable;
export const DocumentHighlight = vscode.DocumentHighlight;
export const DocumentHighlightKind = vscode.DocumentHighlightKind;
export const DocumentLink = vscode.DocumentLink;
export const EndOfLine = vscode.EndOfLine;
export const EventEmitter = vscode.EventEmitter;
export const Hover = vscode.Hover;
export const IndentAction = vscode.IndentAction;
export const Location = vscode.Location;
export const OverviewRulerLane = vscode.OverviewRulerLane;
export const ParameterInformation = vscode.ParameterInformation;
export const Position = vscode.Position;
export const Range = vscode.Range;
export const Selection = vscode.Selection;
export const SignatureHelp = vscode.SignatureHelp;
export const SignatureInformation = vscode.SignatureInformation;
export const SnippetString = vscode.SnippetString;
export const StatusBarAlignment = vscode.StatusBarAlignment;
export const SymbolInformation = vscode.SymbolInformation;
export const SymbolKind = vscode.SymbolKind;
export const TextDocumentSaveReason = vscode.TextDocumentSaveReason;
export const TextEdit = vscode.TextEdit;
export const TextEditorCursorStyle = vscode.TextEditorCursorStyle;
export const TextEditorLineNumbersStyle = vscode.TextEditorLineNumbersStyle;
export const TextEditorRevealType = vscode.TextEditorRevealType;
export const TextEditorSelectionChangeKind = vscode.TextEditorSelectionChangeKind;
export const Uri = vscode.Uri;
export const ViewColumn = vscode.ViewColumn;
export const WorkspaceEdit = vscode.WorkspaceEdit;
// types
export type CancellationTokenSource = typeof vscode.CancellationTokenSource;
export type CodeLens = typeof vscode.CodeLens;
export type CompletionItem = typeof vscode.CompletionItem;
export type CompletionItemKind = typeof vscode.CompletionItemKind;
export type CompletionList = typeof vscode.CompletionList;
export type Diagnostic = typeof vscode.Diagnostic;
export type DiagnosticSeverity = typeof vscode.DiagnosticSeverity;
export type Disposable = typeof vscode.Disposable;
export type DocumentHighlight = typeof vscode.DocumentHighlight;
export type DocumentHighlightKind = typeof vscode.DocumentHighlightKind;
export type DocumentLink = typeof vscode.DocumentLink;
export type EndOfLine = typeof vscode.EndOfLine;
export type EventEmitter = typeof vscode.EventEmitter;
export type Hover = typeof vscode.Hover;
export type IndentAction = typeof vscode.IndentAction;
export type Location = typeof vscode.Location;
export type OverviewRulerLane = typeof vscode.OverviewRulerLane;
export type ParameterInformation = typeof vscode.ParameterInformation;
export type Position = typeof vscode.Position;
export type Range = typeof vscode.Range;
export type Selection = typeof vscode.Selection;
export type SignatureHelp = typeof vscode.SignatureHelp;
export type SignatureInformation = typeof vscode.SignatureInformation;
export type SnippetString = typeof vscode.SnippetString;
export type StatusBarAlignment = typeof vscode.StatusBarAlignment;
export type SymbolInformation = typeof vscode.SymbolInformation;
export type SymbolKind = typeof vscode.SymbolKind;
export type TextDocumentSaveReason = typeof vscode.TextDocumentSaveReason;
export type TextEdit = typeof vscode.TextEdit;
export type TextEditorCursorStyle = typeof vscode.TextEditorCursorStyle;
export type TextEditorLineNumbersStyle = typeof vscode.TextEditorLineNumbersStyle;
export type TextEditorRevealType = typeof vscode.TextEditorRevealType;
export type TextEditorSelectionChangeKind = typeof vscode.TextEditorSelectionChangeKind;
export type Uri = typeof vscode.Uri;
export type ViewColumn = typeof vscode.ViewColumn;
export type WorkspaceEdit = typeof vscode.WorkspaceEdit;
