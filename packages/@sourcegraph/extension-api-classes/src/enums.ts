/* istanbul ignore file */
// tslint:disable: no-object-literal-type-assertion https://github.com/palantir/tslint/issues/4628

import * as sourcegraph from 'sourcegraph'

export const MarkupKind: typeof sourcegraph.MarkupKind = {
    PlainText: 'plaintext' as sourcegraph.MarkupKind.PlainText,
    Markdown: 'markdown' as sourcegraph.MarkupKind.Markdown,
} as const

export const NotificationType: typeof sourcegraph.NotificationType = {
    Error: 1,
    Warning: 2,
    Info: 3,
    Log: 4,
    Success: 5,
} as const
