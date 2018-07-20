# `decoration` LSP extension

The LSP `decoration` extension allows the client to request decorations for a text document from
the server.

Future plans:

- Supporting server push of decorations (e.g., when another user adds a comment to a line, the server preemptively notifies the client of a new decoration indicating the presence of the comment).

### Text Document Decoration Request

_Request_:

- method: 'textDocument/decoration'
- params: `TextDocumentDecorationParams` defined as follows:

```typescript
interface TextDocumentDecorationParams {
  /**
   * The text document.
   */
  textDocument: TextDocumentIdentifier
}
```

_Response_:

- result: `TextDocumentDecoration[]` where `TextDocumentDecoration` is defined as follows:

```typescript
interface TextDocumentDecoration {
  /**
   * The range that the decoration applies to.
   */
  range: Range

  /**
   * Whether the decoration applies to the whole line(s) in the range.
   */
  isWholeLine?: boolean

  /**
   * The background color of the decoration, as a value that is both a valid CSS <color> value and
   * a valid VS Code decoration color. TODO: Specify this better as we implement support for more
   * editors (that are not all Electron-based).
   */
  backgroundColor?: string
}
```

- error: code and message set in case an exception occurs during the request.
