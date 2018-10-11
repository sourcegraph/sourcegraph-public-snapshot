// Package jsonx is an extended JSON library for Go. It is highly tolerant of
// errors, and it supports trailing commas and comments (`//` and `/* ... */`).
//
// It is ported from [Visual Studio Code's](https://github.com/Microsoft/vscode)
// comment-aware JSON parsing and editing APIs in TypeScript.
package jsonx

//go:generate stringer -type=ParseErrorCode,ScanErrorCode,SyntaxKind,NodeType -output=json_stringer.go
