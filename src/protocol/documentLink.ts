import { DocumentLink, TextDocumentIdentifier } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'
import { TextDocumentRegistrationOptions } from './textDocument'

/**
 * Document link options
 */
export interface DocumentLinkOptions {
    /**
     * Document links have a resolve provider as well.
     */
    resolveProvider?: boolean
}

export interface DocumentLinkParams {
    /**
     * The document to provide document links for.
     */
    textDocument: TextDocumentIdentifier
}

/**
 * Document link registration options
 */
export interface DocumentLinkRegistrationOptions extends TextDocumentRegistrationOptions, DocumentLinkOptions {}

/**
 * A request to provide document links
 */
export namespace DocumentLinkRequest {
    export const type = new RequestType<
        DocumentLinkParams,
        DocumentLink[] | null,
        void,
        DocumentLinkRegistrationOptions
    >('textDocument/documentLink')
}

/**
 * Request to resolve additional information for a given document link. The request's
 * parameter is of type [DocumentLink](#DocumentLink) the response
 * is of type [DocumentLink](#DocumentLink) or a Thenable that resolves to such.
 */
export namespace DocumentLinkResolveRequest {
    export const type = new RequestType<DocumentLink, DocumentLink, void, void>('documentLink/resolve')
}
