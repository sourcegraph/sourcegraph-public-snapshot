package lsp

import (
	"github.com/sourcegraph/go-lsp"
)

type None = lsp.None

type InitializeParams = lsp.InitializeParams

type DocumentURI = lsp.DocumentURI

type ClientCapabilities = lsp.ClientCapabilities

type WorkspaceClientCapabilities = lsp.WorkspaceClientCapabilities

type TextDocumentClientCapabilities = lsp.TextDocumentClientCapabilities

type InitializeResult = lsp.InitializeResult

type InitializeError = lsp.InitializeError

// TextDocumentSyncKind is a DEPRECATED way to describe how text
// document syncing works. Use TextDocumentSyncOptions instead (or the
// Options field of TextDocumentSyncOptionsOrKind if you need to
// support JSON-(un)marshaling both).
type TextDocumentSyncKind = lsp.TextDocumentSyncKind

const (
	TDSKNone        = lsp.TDSKNone
	TDSKFull        = lsp.TDSKFull
	TDSKIncremental = lsp.TDSKIncremental
)

type TextDocumentSyncOptions = lsp.TextDocumentSyncOptions

// TextDocumentSyncOptions holds either a TextDocumentSyncKind or
// TextDocumentSyncOptions. The LSP API allows either to be specified
// in the (ServerCapabilities).TextDocumentSync field.
type TextDocumentSyncOptionsOrKind = lsp.TextDocumentSyncOptionsOrKind

type SaveOptions = lsp.SaveOptions

type ServerCapabilities = lsp.ServerCapabilities

type CompletionOptions = lsp.CompletionOptions

type DocumentOnTypeFormattingOptions = lsp.DocumentOnTypeFormattingOptions

type CodeLensOptions = lsp.CodeLensOptions

type SignatureHelpOptions = lsp.SignatureHelpOptions

type ExecuteCommandOptions = lsp.ExecuteCommandOptions

type ExecuteCommandParams = lsp.ExecuteCommandParams

type CompletionItemKind = lsp.CompletionItemKind

const (
	CIKText          = lsp.CIKText
	CIKMethod        = lsp.CIKMethod
	CIKFunction      = lsp.CIKFunction
	CIKConstructor   = lsp.CIKConstructor
	CIKField         = lsp.CIKField
	CIKVariable      = lsp.CIKVariable
	CIKClass         = lsp.CIKClass
	CIKInterface     = lsp.CIKInterface
	CIKModule        = lsp.CIKModule
	CIKProperty      = lsp.CIKProperty
	CIKUnit          = lsp.CIKUnit
	CIKValue         = lsp.CIKValue
	CIKEnum          = lsp.CIKEnum
	CIKKeyword       = lsp.CIKKeyword
	CIKSnippet       = lsp.CIKSnippet
	CIKColor         = lsp.CIKColor
	CIKFile          = lsp.CIKFile
	CIKReference     = lsp.CIKReference
	CIKFolder        = lsp.CIKFolder
	CIKEnumMember    = lsp.CIKEnumMember
	CIKConstant      = lsp.CIKConstant
	CIKStruct        = lsp.CIKStruct
	CIKEvent         = lsp.CIKEvent
	CIKOperator      = lsp.CIKOperator
	CIKTypeParameter = lsp.CIKTypeParameter
)

type CompletionItem = lsp.CompletionItem

type CompletionList = lsp.CompletionList

type CompletionTriggerKind int

const (
	CTKInvoked          CompletionTriggerKind = 1
	CTKTriggerCharacter                       = 2
)

type InsertTextFormat = lsp.InsertTextFormat

const (
	ITFPlainText = lsp.ITFPlainText
	ITFSnippet   = lsp.ITFSnippet
)

type CompletionContext = lsp.CompletionContext

type CompletionParams = lsp.CompletionParams

type Hover = lsp.Hover

type MarkedString = lsp.MarkedString

// RawMarkedString returns a MarkedString consisting of only a raw
// string (i.e., "foo" instead of {"value":"foo", "language":"bar"}).
func RawMarkedString(s string) MarkedString {
	return lsp.RawMarkedString(s)
}

type SignatureHelp = lsp.SignatureHelp

type SignatureInformation = lsp.SignatureInformation

type ParameterInformation = lsp.ParameterInformation

type ReferenceContext = lsp.ReferenceContext

type ReferenceParams = lsp.ReferenceParams

type DocumentHighlightKind = lsp.DocumentHighlightKind

const (
	Text  = lsp.Text
	Read  = lsp.Read
	Write = lsp.Write
)

type DocumentHighlight = lsp.DocumentHighlight

type DocumentSymbolParams = lsp.DocumentSymbolParams

type SymbolKind = lsp.SymbolKind

// The SymbolKind values are defined at https://microsoft.github.io/language-server-protocol/specification.
const (
	SKFile          = lsp.SKFile
	SKModule        = lsp.SKModule
	SKNamespace     = lsp.SKNamespace
	SKPackage       = lsp.SKPackage
	SKClass         = lsp.SKClass
	SKMethod        = lsp.SKMethod
	SKProperty      = lsp.SKProperty
	SKField         = lsp.SKField
	SKConstructor   = lsp.SKConstructor
	SKEnum          = lsp.SKEnum
	SKInterface     = lsp.SKInterface
	SKFunction      = lsp.SKFunction
	SKVariable      = lsp.SKVariable
	SKConstant      = lsp.SKConstant
	SKString        = lsp.SKString
	SKNumber        = lsp.SKNumber
	SKBoolean       = lsp.SKBoolean
	SKArray         = lsp.SKArray
	SKObject        = lsp.SKObject
	SKKey           = lsp.SKKey
	SKNull          = lsp.SKNull
	SKEnumMember    = lsp.SKEnumMember
	SKStruct        = lsp.SKStruct
	SKEvent         = lsp.SKEvent
	SKOperator      = lsp.SKOperator
	SKTypeParameter = lsp.SKTypeParameter
)

type SymbolInformation = lsp.SymbolInformation

type WorkspaceSymbolParams = lsp.WorkspaceSymbolParams

type ConfigurationParams = lsp.ConfigurationParams

type ConfigurationItem = lsp.ConfigurationItem

type ConfigurationResult = lsp.ConfigurationResult

type CodeActionContext = lsp.CodeActionContext

type CodeActionParams = lsp.CodeActionParams

type CodeLensParams = lsp.CodeLensParams

type CodeLens = lsp.CodeLens

type DocumentFormattingParams = lsp.DocumentFormattingParams

type FormattingOptions = lsp.FormattingOptions

type RenameParams = lsp.RenameParams

type DidOpenTextDocumentParams = lsp.DidOpenTextDocumentParams

type DidChangeTextDocumentParams = lsp.DidChangeTextDocumentParams

type TextDocumentContentChangeEvent = lsp.TextDocumentContentChangeEvent

type DidCloseTextDocumentParams = lsp.DidCloseTextDocumentParams

type DidSaveTextDocumentParams = lsp.DidSaveTextDocumentParams

type MessageType = lsp.MessageType

const (
	MTError   = lsp.MTError
	MTWarning = lsp.MTWarning
	Info      = lsp.Info
	Log       = lsp.Log
)

type ShowMessageParams = lsp.ShowMessageParams

type MessageActionItem = lsp.MessageActionItem

type ShowMessageRequestParams = lsp.ShowMessageRequestParams

type LogMessageParams = lsp.LogMessageParams

type DidChangeConfigurationParams = lsp.DidChangeConfigurationParams

type FileChangeType = lsp.FileChangeType

const (
	Created = lsp.Created
	Changed = lsp.Changed
	Deleted = lsp.Deleted
)

type FileEvent = lsp.FileEvent

type DidChangeWatchedFilesParams = lsp.DidChangeWatchedFilesParams

type PublishDiagnosticsParams = lsp.PublishDiagnosticsParams

type DocumentRangeFormattingParams = lsp.DocumentRangeFormattingParams

type DocumentOnTypeFormattingParams = lsp.DocumentOnTypeFormattingParams

type CancelParams = lsp.CancelParams
