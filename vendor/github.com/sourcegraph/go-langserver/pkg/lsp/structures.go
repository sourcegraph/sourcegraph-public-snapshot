package lsp

import (
	"github.com/sourcegraph/go-lsp"
)

type Position = lsp.Position

type Range = lsp.Range

type Location = lsp.Location

type Diagnostic = lsp.Diagnostic

type DiagnosticSeverity = lsp.DiagnosticSeverity

const (
	Error       = lsp.Error
	Warning     = lsp.Warning
	Information = lsp.Information
	Hint        = lsp.Hint
)

type Command = lsp.Command

type TextEdit = lsp.TextEdit

type WorkspaceEdit = lsp.WorkspaceEdit

type TextDocumentIdentifier = lsp.TextDocumentIdentifier

type TextDocumentItem = lsp.TextDocumentItem

type VersionedTextDocumentIdentifier = lsp.VersionedTextDocumentIdentifier

type TextDocumentPositionParams = lsp.TextDocumentPositionParams
