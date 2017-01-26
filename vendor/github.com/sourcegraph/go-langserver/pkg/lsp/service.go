package lsp

import (
	"bytes"
	"encoding/json"
)

type None struct{}

type InitializeParams struct {
	ProcessID             int                `json:"processId,omitempty"`
	RootPath              string             `json:"rootPath,omitempty"`
	InitializationOptions interface{}        `json:"initializationOptions,omitempty"`
	Capabilities          ClientCapabilities `json:"capabilities"`
}

type ClientCapabilities struct {
	// Below are Sourcegraph extensions. They do not live in lspext since
	// they are extending the field InitializeParams.Capabilities

	// XFilesProvider indicates the client provides support for
	// workspace/xfiles. This is a Sourcegraph extension.
	XFilesProvider bool `json:"xfilesProvider,omitempty"`

	// XContentProvider indicates the client provides support for
	// textDocument/xcontent. This is a Sourcegraph extension.
	XContentProvider bool `json:"xcontentProvider,omitempty"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities,omitempty"`
}

type InitializeError struct {
	Retry bool `json:"retry"`
}

type TextDocumentSyncKind int

const (
	TDSKNone        TextDocumentSyncKind = 0
	TDSKFull        TextDocumentSyncKind = 1
	TDSKIncremental TextDocumentSyncKind = 2
)

type TextDocumentSyncOptions struct {
	OpenClose         bool                 `json:"openClose,omitempty"`
	Change            TextDocumentSyncKind `json:"change"`
	WillSave          bool                 `json:"willSave,omitempty"`
	WillSaveWaitUntil bool                 `json:"willSaveWaitUntil,omitempty"`
	Save              *SaveOptions         `json:"save,omitempty"`
}

// TextDocumentSyncOptions holds either a TextDocumentSyncKind or
// TextDocumentSyncOptions. The LSP API allows either to be specified
// in the (ServerCapabilities).TextDocumentSync field.
type TextDocumentSyncOptionsOrKind struct {
	Kind    *TextDocumentSyncKind
	Options *TextDocumentSyncOptions
}

// MarshalJSON implements json.Marshaler.
func (v TextDocumentSyncOptionsOrKind) MarshalJSON() ([]byte, error) {
	if v.Kind != nil {
		return json.Marshal(v.Kind)
	}
	return json.Marshal(v.Options)
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *TextDocumentSyncOptionsOrKind) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*v = TextDocumentSyncOptionsOrKind{}
		return nil
	}
	var kind TextDocumentSyncKind
	if err := json.Unmarshal(data, &kind); err == nil {
		// Create equivalent TextDocumentSyncOptions using the same
		// logic as in vscode-languageclient. Also set the Kind field
		// so that JSON-marshaling and unmarshaling are inverse
		// operations (for backward compatibility, preserving the
		// original input but accepting both).
		*v = TextDocumentSyncOptionsOrKind{
			Options: &TextDocumentSyncOptions{OpenClose: true, Change: kind},
			Kind:    &kind,
		}
		return nil
	}
	var tmp TextDocumentSyncOptions
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*v = TextDocumentSyncOptionsOrKind{Options: &tmp}
	return nil
}

type SaveOptions struct {
	IncludeText bool `json:"includeText"`
}

type ServerCapabilities struct {
	TextDocumentSync                 TextDocumentSyncOptionsOrKind    `json:"textDocumentSync,omitempty"`
	HoverProvider                    bool                             `json:"hoverProvider,omitempty"`
	CompletionProvider               *CompletionOptions               `json:"completionProvider,omitempty"`
	SignatureHelpProvider            *SignatureHelpOptions            `json:"signatureHelpProvider,omitempty"`
	DefinitionProvider               bool                             `json:"definitionProvider,omitempty"`
	ReferencesProvider               bool                             `json:"referencesProvider,omitempty"`
	DocumentHighlightProvider        bool                             `json:"documentHighlightProvider,omitempty"`
	DocumentSymbolProvider           bool                             `json:"documentSymbolProvider,omitempty"`
	WorkspaceSymbolProvider          bool                             `json:"workspaceSymbolProvider,omitempty"`
	CodeActionProvider               bool                             `json:"codeActionProvider,omitempty"`
	CodeLensProvider                 *CodeLensOptions                 `json:"codeLensProvider,omitempty"`
	DocumentFormattingProvider       bool                             `json:"documentFormattingProvider,omitempty"`
	DocumentRangeFormattingProvider  bool                             `json:"documentRangeFormattingProvider,omitempty"`
	DocumentOnTypeFormattingProvider *DocumentOnTypeFormattingOptions `json:"documentOnTypeFormattingProvider,omitempty"`
	RenameProvider                   bool                             `json:"renameProvider,omitempty"`

	// XWorkspaceReferencesProvider indicates the server provides support for
	// xworkspace/references. This is a Sourcegraph extension.
	XWorkspaceReferencesProvider bool `json:"xworkspaceReferencesProvider,omitempty"`

	// XDefinitionProvider indicates the server provides support for
	// textDocument/xdefinition. This is a Sourcegraph extension.
	XDefinitionProvider bool `json:"xdefinitionProvider,omitempty"`

	// XWorkspaceSymbolByProperties indicates the server provides support for
	// querying symbols by properties with WorkspaceSymbolParams.symbol. This
	// is a Sourcegraph extension.
	XWorkspaceSymbolByProperties bool `json:"xworkspaceSymbolByProperties,omitempty"`
}

type CompletionOptions struct {
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type DocumentOnTypeFormattingOptions struct {
	FirstTriggerCharacter string   `json:"firstTriggerCharacter"`
	MoreTriggerCharacter  []string `json:"moreTriggerCharacter,omitempty"`
}

type CodeLensOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

type SignatureHelpOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type CompletionItemKind int

const (
	CIKText        CompletionItemKind = 1
	CIKMethod                         = 2
	CIKFunction                       = 3
	CIKConstructor                    = 4
	CIKField                          = 5
	CIKVariable                       = 6
	CIKClass                          = 7
	CIKInterface                      = 8
	CIKModule                         = 9
	CIKProperty                       = 10
	CIKUnit                           = 11
	CIKValue                          = 12
	CIKEnum                           = 13
	CIKKeyword                        = 14
	CIKSnippet                        = 15
	CIKColor                          = 16
	CIKFile                           = 17
	CIKReference                      = 18
)

type CompletionItem struct {
	Label         string      `json:"label"`
	Kind          int         `json:"kind,omitempty"`
	Detail        string      `json:"detail,omitempty"`
	Documentation string      `json:"documentation,omitempty"`
	SortText      string      `json:"sortText,omitempty"`
	FilterText    string      `json:"filterText,omitempty"`
	InsertText    string      `json:"insertText,omitempty"`
	TextEdit      TextEdit    `json:"textEdit,omitempty"`
	Data          interface{} `json:"data,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type Hover struct {
	Contents []MarkedString `json:"contents,omitempty"`
	Range    Range          `json:"range"`
}

type MarkedString struct {
	Language string `json:"language"`
	Value    string `json:"value"`
}

type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature"`
	ActiveParameter int                    `json:"activeParameter"`
}

type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation string                 `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters,omitempty"`
}

type ParameterInformation struct {
	Label         string `json:"label"`
	Documentation string `json:"documentation,omitempty"`
}

type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`

	// Sourcegraph extension
	XLimit int `json:"xlimit,omitempty"`
}

type ReferenceParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

type DocumentHighlightKind int

const (
	Text  DocumentHighlightKind = 1
	Read                        = 2
	Write                       = 3
)

type DocumentHighlight struct {
	Range Range `json:"range"`
	Kind  int   `json:"kind,omitempty"`
}

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type SymbolKind int

const (
	SKFile        SymbolKind = 1
	SKModule      SymbolKind = 2
	SKNamespace   SymbolKind = 3
	SKPackage     SymbolKind = 4
	SKClass       SymbolKind = 5
	SKMethod      SymbolKind = 6
	SKProperty    SymbolKind = 7
	SKField       SymbolKind = 8
	SKConstructor SymbolKind = 9
	SKEnum        SymbolKind = 10
	SKInterface   SymbolKind = 11
	SKFunction    SymbolKind = 12
	SKVariable    SymbolKind = 13
	SKConstant    SymbolKind = 14
	SKString      SymbolKind = 15
	SKNumber      SymbolKind = 16
	SKBoolean     SymbolKind = 17
	SKArray       SymbolKind = 18
)

func (s SymbolKind) String() string {
	return symbolKindName[s]
}

var symbolKindName = map[SymbolKind]string{
	SKFile:        "file",
	SKModule:      "module",
	SKNamespace:   "namespace",
	SKPackage:     "package",
	SKClass:       "class",
	SKMethod:      "method",
	SKProperty:    "property",
	SKField:       "field",
	SKConstructor: "constructor",
	SKEnum:        "enum",
	SKInterface:   "interface",
	SKFunction:    "function",
	SKVariable:    "variable",
	SKConstant:    "constant",
	SKString:      "string",
	SKNumber:      "number",
	SKBoolean:     "boolean",
	SKArray:       "array",
}

type SymbolInformation struct {
	Name          string     `json:"name"`
	Kind          SymbolKind `json:"kind"`
	Location      Location   `json:"location"`
	ContainerName string     `json:"containerName,omitempty"`
}

type WorkspaceSymbolParams struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

type CodeLensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type CodeLens struct {
	Range   Range       `json:"range"`
	Command Command     `json:"command,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

type FormattingOptions struct {
	TabSize      int    `json:"tabSize"`
	InsertSpaces bool   `json:"insertSpaces"`
	Key          string `json:"key"`
}

type RenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	NewName      string                 `json:"newName"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitEmpty"`
	RangeLength uint   `json:"rangeLength,omitEmpty"`
	Text        string `json:"text"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type MessageType int

const (
	MTError   MessageType = 1
	MTWarning             = 2
	Info                  = 3
	Log                   = 4
)

type ShowMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

type MessageActionItem struct {
	Title string `json:"title"`
}

type ShowMessageRequestParams struct {
	Type    MessageType         `json:"type"`
	Message string              `json:"message"`
	Actions []MessageActionItem `json:"actions"`
}

type LogMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

type DidChangeConfigurationParams struct {
	Settings interface{} `json:"settings"`
}

type FileChangeType int

const (
	Created FileChangeType = 1
	Changed                = 2
	Deleted                = 3
)

type FileEvent struct {
	URI  string `json:"uri"`
	Type int    `json:"type"`
}

type DidChangeWatchedFilesParams struct {
	Changes []FileEvent `json:"changes"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type DocumentRangeFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Options      FormattingOptions      `json:"options"`
}

type DocumentOnTypeFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Ch           string                 `json:"ch"`
	Options      FormattingOptions      `json:"formattingOptions"`
}

type CancelParams struct {
	ID ID `json:"id"`
}
