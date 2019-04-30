import * as sourcegraph from 'sourcegraph'

export const MarkupKind: typeof sourcegraph.MarkupKind = {
    PlainText: 'plaintext' as sourcegraph.MarkupKind.PlainText,
    Markdown: 'markdown' as sourcegraph.MarkupKind.Markdown,
}
