import type { DocumentSelector } from 'sourcegraph'

import type { Position } from '@sourcegraph/extension-api-types'

/**
 * The URI scheme for the resources that hold the body of comments (such as comments on a GitHub
 * issue).
 */
export const COMMENT_URI_SCHEME = 'comment'

/**
 * The URI scheme for the resources that hold the body of snippets.
 */
export const SNIPPET_URI_SCHEME = 'snippet'

/**
 * A literal to identify a text document in the client.
 */
export interface TextDocumentIdentifier {
    /**
     * The text document's URI.
     */
    uri: string
}

/**
 * A parameter literal used in requests to pass a text document and a position inside that
 * document.
 */
export interface TextDocumentPositionParameters {
    /**
     * The text document.
     */
    textDocument: TextDocumentIdentifier

    /**
     * The position inside the text document.
     */
    position: Position
}

/**
 * General text document registration options.
 */
export interface TextDocumentRegistrationOptions {
    /**
     * A document selector to identify the scope of the registration.
     */
    documentSelector: DocumentSelector
}
