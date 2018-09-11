import { DocumentSelector, MarkupKind } from 'sourcegraph'
import { TextDocumentIdentifier, TextDocumentItem } from '../client/types/textDocument'
import { Position } from './plainTypes'

/**
 * A parameter literal used in requests to pass a text document and a position inside that
 * document.
 */
export interface TextDocumentPositionParams {
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
 * Text document specific client capabilities.
 */
export interface TextDocumentClientCapabilities {
    /**
     * Defines which synchronization capabilities the client supports.
     */
    synchronization?: {
        /**
         * Whether text document synchronization supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }

    /**
     * Capabilities specific to the `textDocument/hover`
     */
    hover?: {
        /**
         * Whether hover supports dynamic registration.
         */
        dynamicRegistration?: boolean

        /**
         * Client supports the follow content formats for the content
         * property. The order describes the preferred format of the client.
         */
        contentFormat?: MarkupKind[]
    }

    /**
     * Capabilities specific to the `textDocument/references`
     */
    references?: {
        /**
         * Whether references supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }

    /**
     * Capabilities specific to the `textDocument/definition`
     */
    definition?: {
        /**
         * Whether definition supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }
}

/**
 * General text document registration options.
 */
export interface TextDocumentRegistrationOptions {
    /**
     * A document selector to identify the scope of the registration. If set to null
     * the document selector provided on the client side will be used.
     */
    documentSelector: DocumentSelector | null

    /**
     * ID of the extension that registers the provider.
     */
    extensionID: string
}

export interface TextDocumentSyncOptions {
    /**
     * Open and close notifications are sent to the server.
     */
    openClose?: boolean
}

/**
 * The parameters send in a open text document notification
 */
export interface DidOpenTextDocumentParams {
    /**
     * The document that was opened.
     */
    textDocument: TextDocumentItem
}

/**
 * The document open notification is sent from the client to the server to signal
 * newly opened text documents. The document's truth is now managed by the client
 * and the server must not try to read the document's truth using the document's
 * uri. Open in this sense means it is managed by the client. It doesn't necessarily
 * mean that its content is presented in an editor. An open notification must not
 * be sent more than once without a corresponding close notification send before.
 * This means open and close notification must be balanced and the max open count
 * is one.
 */
export namespace DidOpenTextDocumentNotification {
    export const type = 'textDocument/didOpen'
}

/**
 * The parameters send in a close text document notification
 */
export interface DidCloseTextDocumentParams {
    /**
     * The document that was closed.
     */
    textDocument: TextDocumentIdentifier
}

/**
 * The document close notification is sent from the client to the server when
 * the document got closed in the client. The document's truth now exists where
 * the document's uri points to (e.g. if the document's uri is a file uri the
 * truth now exists on disk). As with the open notification the close notification
 * is about managing the document's content. Receiving a close notification
 * doesn't mean that the document was open in an editor before. A close
 * notification requires a previous open notification to be sent.
 */
export namespace DidCloseTextDocumentNotification {
    export const type = 'textDocument/didClose'
}
