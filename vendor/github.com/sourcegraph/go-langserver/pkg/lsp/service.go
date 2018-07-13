package lsp

import (
	"bytes"
	"encoding/json"
	"strings"
)

type None struct{}

type InitializeParams struct {
	ProcessID int `json:"processId,omitempty"`

	// RootPath is DEPRECATED in favor of the RootURI field.
	RootPath string `json:"rootPath,omitempty"`

	RootURI               DocumentURI        `json:"rootUri,omitempty"`
	InitializationOptions interface{}        `json:"initializationOptions,omitempty"`
	Capabilities          ClientCapabilities `json:"capabilities"`
}

// Root returns the RootURI if set, or otherwise the RootPath with 'file://' prepended.
func (p *InitializeParams) Root() DocumentURI {
	if p.RootURI != "" {
		return p.RootURI
	}
	if strings.HasPrefix(p.RootPath, "file://") {
		return DocumentURI(p.RootPath)
	}
	return DocumentURI("file://" + p.RootPath)
}

type DocumentURI string

type ClientCapabilities struct {
	Workspace    WorkspaceClientCapabilities    `json:"workspace,omitempty"`
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
	Experimental interface{}                    `json:"experimental,omitempty"`

	// Below are Sourcegraph extensions. They do not live in lspext since
	// they are extending the field InitializeParams.Capabilities

	// XFilesProvider indicates the client provides support for
	// workspace/xfiles. This is a Sourcegraph extension.
	XFilesProvider bool `json:"xfilesProvider,omitempty"`

	// XContentProvider indicates the client provides support for
	// textDocument/xcontent. This is a Sourcegraph extension.
	XContentProvider bool `json:"xcontentProvider,omitempty"`

	// XCacheProvider indicates the client provides support for cache/get
	// and cache/set.
	XCacheProvider bool `json:"xcacheProvider,omitempty"`
}

type WorkspaceClientCapabilities struct{}

type TextDocumentClientCapabilities struct {
	Completion struct {
		CompletionItemKind struct {
			ValueSet []CompletionItemKind `json:"valueSet,omitempty"`
		} `json:"completionItemKind,omitempty"`
		CompletionItem struct {
			SnippetSupport bool `json:"snippetSupport,omitempty"`
		} `json:"completionItem,omitempty"`
	} `json:"completion,omitempty"`

	Implementation *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"implementation,omitempty"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities,omitempty"`
}

type InitializeError struct {
	Retry bool `json:"retry"`
}

// TextDocumentSyncKind is a DEPRECATED way to describe how text
// document syncing works. Use TextDocumentSyncOptions instead (or the
// Options field of TextDocumentSyncOptionsOrKind if you need to
// support JSON-(un)marshaling both).
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
func (v *TextDocumentSyncOptionsOrKind) MarshalJSON() ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
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
	TextDocumentSync                 *TextDocumentSyncOptionsOrKind   `json:"textDocumentSync,omitempty"`
	HoverProvider                    bool                             `json:"hoverProvider,omitempty"`
	CompletionProvider               *CompletionOptions               `json:"completionProvider,omitempty"`
	SignatureHelpProvider            *SignatureHelpOptions            `json:"signatureHelpProvider,omitempty"`
	DefinitionProvider               bool                             `json:"definitionProvider,omitempty"`
	TypeDefinitionProvider           bool                             `json:"typeDefinitionProvider,omitempty"`
	ReferencesProvider               bool                             `json:"referencesProvider,omitempty"`
	DocumentHighlightProvider        bool                             `json:"documentHighlightProvider,omitempty"`
	DocumentSymbolProvider           bool                             `json:"documentSymbolProvider,omitempty"`
	WorkspaceSymbolProvider          bool                             `json:"workspaceSymbolProvider,omitempty"`
	ImplementationProvider           bool                             `json:"implementationProvider,omitempty"`
	CodeActionProvider               bool                             `json:"codeActionProvider,omitempty"`
	CodeLensProvider                 *CodeLensOptions                 `json:"codeLensProvider,omitempty"`
	DocumentFormattingProvider       bool                             `json:"documentFormattingProvider,omitempty"`
	DocumentRangeFormattingProvider  bool                             `json:"documentRangeFormattingProvider,omitempty"`
	DocumentOnTypeFormattingProvider *DocumentOnTypeFormattingOptions `json:"documentOnTypeFormattingProvider,omitempty"`
	RenameProvider                   bool                             `json:"renameProvider,omitempty"`
	ExecuteCommandProvider           *ExecuteCommandOptions           `json:"executeCommandProvider,omitempty"`

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

	Experimental interface{} `json:"experimental,omitempty"`
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

type ExecuteCommandOptions struct {
	Commands []string `json:"commands"`
}

type ExecuteCommandParams struct {
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

type CompletionItemKind int

const (
	_ CompletionItemKind = iota
	CIKText
	CIKMethod
	CIKFunction
	CIKConstructor
	CIKField
	CIKVariable
	CIKClass
	CIKInterface
	CIKModule
	CIKProperty
	CIKUnit
	CIKValue
	CIKEnum
	CIKKeyword
	CIKSnippet
	CIKColor
	CIKFile
	CIKReference
	CIKFolder
	CIKEnumMember
	CIKConstant
	CIKStruct
	CIKEvent
	CIKOperator
	CIKTypeParameter
)

func (c CompletionItemKind) String() string {
	return completionItemKindName[c]
}

var completionItemKindName = map[CompletionItemKind]string{
	CIKText:          "text",
	CIKMethod:        "method",
	CIKFunction:      "function",
	CIKConstructor:   "constructor",
	CIKField:         "field",
	CIKVariable:      "variable",
	CIKClass:         "class",
	CIKInterface:     "interface",
	CIKModule:        "module",
	CIKProperty:      "property",
	CIKUnit:          "unit",
	CIKValue:         "value",
	CIKEnum:          "enum",
	CIKKeyword:       "keyword",
	CIKSnippet:       "snippet",
	CIKColor:         "color",
	CIKFile:          "file",
	CIKReference:     "reference",
	CIKFolder:        "folder",
	CIKEnumMember:    "enumMember",
	CIKConstant:      "constant",
	CIKStruct:        "struct",
	CIKEvent:         "event",
	CIKOperator:      "operator",
	CIKTypeParameter: "typeParameter",
}

type CompletionItem struct {
	Label            string             `json:"label"`
	Kind             CompletionItemKind `json:"kind,omitempty"`
	Detail           string             `json:"detail,omitempty"`
	Documentation    string             `json:"documentation,omitempty"`
	SortText         string             `json:"sortText,omitempty"`
	FilterText       string             `json:"filterText,omitempty"`
	InsertText       string             `json:"insertText,omitempty"`
	InsertTextFormat InsertTextFormat   `json:"insertTextFormat,omitempty"`
	TextEdit         *TextEdit          `json:"textEdit,omitempty"`
	Data             interface{}        `json:"data,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type CompletionTriggerKind int

const (
	CTKInvoked          CompletionTriggerKind = 1
	CTKTriggerCharacter                       = 2
)

type InsertTextFormat int

const (
	ITFPlainText InsertTextFormat = 1
	ITFSnippet                    = 2
)

type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
}

type CompletionParams struct {
	TextDocumentPositionParams
	Context CompletionContext `json:"context,omitempty"`
}

type Hover struct {
	Contents []MarkedString `json:"contents"`
	Range    *Range         `json:"range,omitempty"`
}

type hover Hover

func (h Hover) MarshalJSON() ([]byte, error) {
	if h.Contents == nil {
		return json.Marshal(hover{
			Contents: []MarkedString{},
			Range:    h.Range,
		})
	}
	return json.Marshal(hover(h))
}

type MarkedString markedString

type markedString struct {
	Language string `json:"language"`
	Value    string `json:"value"`

	isRawString bool
}

func (m *MarkedString) UnmarshalJSON(data []byte) error {
	if d := strings.TrimSpace(string(data)); len(d) > 0 && d[0] == '"' {
		// Raw string
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		m.Value = s
		m.isRawString = true
		return nil
	}
	// Language string
	ms := (*markedString)(m)
	return json.Unmarshal(data, ms)
}

func (m MarkedString) MarshalJSON() ([]byte, error) {
	if m.isRawString {
		return json.Marshal(m.Value)
	}
	return json.Marshal((markedString)(m))
}

// RawMarkedString returns a MarkedString consisting of only a raw
// string (i.e., "foo" instead of {"value":"foo", "language":"bar"}).
func RawMarkedString(s string) MarkedString {
	return MarkedString{Value: s, isRawString: true}
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

// The SymbolKind values are defined at https://microsoft.github.io/language-server-protocol/specification.
const (
	SKFile          SymbolKind = 1
	SKModule        SymbolKind = 2
	SKNamespace     SymbolKind = 3
	SKPackage       SymbolKind = 4
	SKClass         SymbolKind = 5
	SKMethod        SymbolKind = 6
	SKProperty      SymbolKind = 7
	SKField         SymbolKind = 8
	SKConstructor   SymbolKind = 9
	SKEnum          SymbolKind = 10
	SKInterface     SymbolKind = 11
	SKFunction      SymbolKind = 12
	SKVariable      SymbolKind = 13
	SKConstant      SymbolKind = 14
	SKString        SymbolKind = 15
	SKNumber        SymbolKind = 16
	SKBoolean       SymbolKind = 17
	SKArray         SymbolKind = 18
	SKObject        SymbolKind = 19
	SKKey           SymbolKind = 20
	SKNull          SymbolKind = 21
	SKEnumMember    SymbolKind = 22
	SKStruct        SymbolKind = 23
	SKEvent         SymbolKind = 24
	SKOperator      SymbolKind = 25
	SKTypeParameter SymbolKind = 26
)

func (s SymbolKind) String() string {
	return symbolKindName[s]
}

var symbolKindName = map[SymbolKind]string{
	SKFile:          "File",
	SKModule:        "Module",
	SKNamespace:     "Namespace",
	SKPackage:       "Package",
	SKClass:         "Class",
	SKMethod:        "Method",
	SKProperty:      "Property",
	SKField:         "Field",
	SKConstructor:   "Constructor",
	SKEnum:          "Enum",
	SKInterface:     "Interface",
	SKFunction:      "Function",
	SKVariable:      "Variable",
	SKConstant:      "Constant",
	SKString:        "String",
	SKNumber:        "Number",
	SKBoolean:       "Boolean",
	SKArray:         "Array",
	SKObject:        "Object",
	SKKey:           "Key",
	SKNull:          "Null",
	SKEnumMember:    "EnumMember",
	SKStruct:        "Struct",
	SKEvent:         "Event",
	SKOperator:      "Operator",
	SKTypeParameter: "TypeParameter",
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

type ConfigurationParams struct {
	Items []ConfigurationItem `json:"items"`
}

type ConfigurationItem struct {
	ScopeURI string `json:"scopeUri,omitempty"`
	Section  string `json:"section,omitempty"`
}

type ConfigurationResult []interface{}

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
	URI  DocumentURI `json:"uri"`
	Type int         `json:"type"`
}

type DidChangeWatchedFilesParams struct {
	Changes []FileEvent `json:"changes"`
}

type PublishDiagnosticsParams struct {
	URI         DocumentURI  `json:"uri"`
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
