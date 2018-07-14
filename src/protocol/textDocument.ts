import {
    CodeActionKind,
    CompletionItemKind,
    MarkupKind,
    Position,
    SymbolKind,
    TextDocumentContentChangeEvent,
    TextDocumentIdentifier,
    TextDocumentItem,
    TextDocumentSaveReason,
    TextEdit,
    VersionedTextDocumentIdentifier,
} from 'vscode-languageserver-types'
import { NotificationType, RequestType } from '../jsonrpc2/messages'
import { DocumentSelector } from '../types/document'
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

        /**
         * The client supports sending will save notifications.
         */
        willSave?: boolean

        /**
         * The client supports sending a will save request and
         * waits for a response providing text edits which will
         * be applied to the document before it is saved.
         */
        willSaveWaitUntil?: boolean

        /**
         * The client supports did save notifications.
         */
        didSave?: boolean
    }

    /**
     * Capabilities specific to the `textDocument/completion`
     */
    completion?: {
        /**
         * Whether completion supports dynamic registration.
         */
        dynamicRegistration?: boolean

        /**
         * The client supports the following `CompletionItem` specific
         * capabilities.
         */
        completionItem?: {
            /**
             * Client supports snippets as insert text.
             *
             * A snippet can define tab stops and placeholders with `$1`, `$2`
             * and `${3:foo}`. `$0` defines the final tab stop, it defaults to
             * the end of the snippet. Placeholders with equal identifiers are linked,
             * that is typing in one will update others too.
             */
            snippetSupport?: boolean

            /**
             * Client supports commit characters on a completion item.
             */
            commitCharactersSupport?: boolean

            /**
             * Client supports the follow content formats for the documentation
             * property. The order describes the preferred format of the client.
             */
            documentationFormat?: MarkupKind[]

            /**
             * Client supports the deprecated property on a completion item.
             */
            deprecatedSupport?: boolean

            /**
             * Client supports the preselect property on a completion item.
             */
            preselectSupport?: boolean
        }

        completionItemKind?: {
            /**
             * The completion item kind values the client supports. When this
             * property exists the client also guarantees that it will
             * handle values outside its set gracefully and falls back
             * to a default value when unknown.
             *
             * If this property is not present the client only supports
             * the completion items kinds from `Text` to `Reference` as defined in
             * the initial version of the protocol.
             */
            valueSet?: CompletionItemKind[]
        }

        /**
         * The client supports to send additional context information for a
         * `textDocument/completion` requestion.
         */
        contextSupport?: boolean
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
     * Capabilities specific to the `textDocument/signatureHelp`
     */
    signatureHelp?: {
        /**
         * Whether signature help supports dynamic registration.
         */
        dynamicRegistration?: boolean

        /**
         * The client supports the following `SignatureInformation`
         * specific properties.
         */
        signatureInformation?: {
            /**
             * Client supports the follow content formats for the documentation
             * property. The order describes the preferred format of the client.
             */
            documentationFormat?: MarkupKind[]
        }
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
     * Capabilities specific to the `textDocument/documentHighlight`
     */
    documentHighlight?: {
        /**
         * Whether document highlight supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }

    /**
     * Capabilities specific to the `textDocument/documentSymbol`
     */
    documentSymbol?: {
        /**
         * Whether document symbol supports dynamic registration.
         */
        dynamicRegistration?: boolean

        /**
         * Specific capabilities for the `SymbolKind`.
         */
        symbolKind?: {
            /**
             * The symbol kind values the client supports. When this
             * property exists the client also guarantees that it will
             * handle values outside its set gracefully and falls back
             * to a default value when unknown.
             *
             * If this property is not present the client only supports
             * the symbol kinds from `File` to `Array` as defined in
             * the initial version of the protocol.
             */
            valueSet?: SymbolKind[]
        }
    }

    /**
     * Capabilities specific to the `textDocument/formatting`
     */
    formatting?: {
        /**
         * Whether formatting supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }

    /**
     * Capabilities specific to the `textDocument/rangeFormatting`
     */
    rangeFormatting?: {
        /**
         * Whether range formatting supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }

    /**
     * Capabilities specific to the `textDocument/onTypeFormatting`
     */
    onTypeFormatting?: {
        /**
         * Whether on type formatting supports dynamic registration.
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

    /**
     * Capabilities specific to the `textDocument/codeAction`
     */
    codeAction?: {
        /**
         * Whether code action supports dynamic registration.
         */
        dynamicRegistration?: boolean

        /**
         * The client support code action literals as a valid
         * response of the `textDocument/codeAction` request.
         */
        codeActionLiteralSupport?: {
            /**
             * The code action kind is support with the following value
             * set.
             */
            codeActionKind: {
                /**
                 * The code action kind values the client supports. When this
                 * property exists the client also guarantees that it will
                 * handle values outside its set gracefully and falls back
                 * to a default value when unknown.
                 */
                valueSet: CodeActionKind[]
            }
        }
    }

    /**
     * Capabilities specific to the `textDocument/codeLens`
     */
    codeLens?: {
        /**
         * Whether code lens supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }

    /**
     * Capabilities specific to the `textDocument/documentLink`
     */
    documentLink?: {
        /**
         * Whether document link supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }

    /**
     * Capabilities specific to the `textDocument/rename`
     */
    rename?: {
        /**
         * Whether rename supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }

    /**
     * Capabilities specific to `textDocument/publishDiagnostics`.
     */
    publishDiagnostics?: {
        /**
         * Whether the clients accepts diagnostics with related information.
         */
        relatedInformation?: boolean
    }
}

/**
 * Defines how the host (editor) should sync
 * document changes to the language server.
 */
export namespace TextDocumentSyncKind {
    /**
     * Documents should not be synced at all.
     */
    export const None = 0

    /**
     * Documents are synced by always sending the full content
     * of the document.
     */
    export const Full = 1

    /**
     * Documents are synced by sending the full content on open.
     * After that only incremental updates to the document are
     * send.
     */
    export const Incremental = 2
}

export type TextDocumentSyncKind = 0 | 1 | 2

/**
 * General text document registration options.
 */
export interface TextDocumentRegistrationOptions {
    /**
     * A document selector to identify the scope of the registration. If set to null
     * the document selector provided on the client side will be used.
     */
    documentSelector: DocumentSelector | null
}

/**
 * Save options.
 */
export interface SaveOptions {
    /**
     * The client is supposed to include the content on save.
     */
    includeText?: boolean
}

export interface TextDocumentSyncOptions {
    /**
     * Open and close notifications are sent to the server.
     */
    openClose?: boolean
    /**
     * Change notifications are sent to the server. See TextDocumentSyncKind.None, TextDocumentSyncKind.Full
     * and TextDocumentSyncKind.Incremental.
     */
    change?: TextDocumentSyncKind
    /**
     * Will save notifications are sent to the server.
     */
    willSave?: boolean
    /**
     * Will save wait until requests are sent to the server.
     */
    willSaveWaitUntil?: boolean
    /**
     * Save notifications are sent to the server.
     */
    save?: SaveOptions
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
    export const type = new NotificationType<DidOpenTextDocumentParams, TextDocumentRegistrationOptions>(
        'textDocument/didOpen'
    )
}

/**
 * The change text document notification's parameters.
 */
export interface DidChangeTextDocumentParams {
    /**
     * The document that did change. The version number points
     * to the version after all provided content changes have
     * been applied.
     */
    textDocument: VersionedTextDocumentIdentifier

    /**
     * The actual content changes. The content changes describe single state changes
     * to the document. So if there are two content changes c1 and c2 for a document
     * in state S10 then c1 move the document to S11 and c2 to S12.
     */
    contentChanges: TextDocumentContentChangeEvent[]
}

/**
 * Describe options to be used when registered for text document change events.
 */
export interface TextDocumentChangeRegistrationOptions extends TextDocumentRegistrationOptions {
    /**
     * How documents are synced to the server.
     */
    syncKind: TextDocumentSyncKind
}

/**
 * The document change notification is sent from the client to the server to signal
 * changes to a text document.
 */
export namespace DidChangeTextDocumentNotification {
    export const type = new NotificationType<DidChangeTextDocumentParams, TextDocumentChangeRegistrationOptions>(
        'textDocument/didChange'
    )
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
    export const type = new NotificationType<DidCloseTextDocumentParams, TextDocumentRegistrationOptions>(
        'textDocument/didClose'
    )
}

/**
 * The parameters send in a save text document notification
 */
export interface DidSaveTextDocumentParams {
    /**
     * The document that was closed.
     */
    textDocument: VersionedTextDocumentIdentifier

    /**
     * Optional the content when saved. Depends on the includeText value
     * when the save notification was requested.
     */
    text?: string
}

/**
 * Save registration options.
 */
export interface TextDocumentSaveRegistrationOptions extends TextDocumentRegistrationOptions, SaveOptions {}

/**
 * The document save notification is sent from the client to the server when
 * the document got saved in the client.
 */
export namespace DidSaveTextDocumentNotification {
    export const type = new NotificationType<DidSaveTextDocumentParams, TextDocumentSaveRegistrationOptions>(
        'textDocument/didSave'
    )
}

/**
 * The parameters send in a will save text document notification.
 */
export interface WillSaveTextDocumentParams {
    /**
     * The document that will be saved.
     */
    textDocument: TextDocumentIdentifier

    /**
     * The 'TextDocumentSaveReason'.
     */
    reason: TextDocumentSaveReason
}

/**
 * A document will save notification is sent from the client to the server before
 * the document is actually saved.
 */
export namespace WillSaveTextDocumentNotification {
    export const type = new NotificationType<WillSaveTextDocumentParams, TextDocumentRegistrationOptions>(
        'textDocument/willSave'
    )
}

/**
 * A document will save request is sent from the client to the server before
 * the document is actually saved. The request can return an array of TextEdits
 * which will be applied to the text document before it is saved. Please note that
 * clients might drop results if computing the text edits took too long or if a
 * server constantly fails on this request. This is done to keep the save fast and
 * reliable.
 */
export namespace WillSaveTextDocumentWaitUntilRequest {
    export const type = new RequestType<
        WillSaveTextDocumentParams,
        TextEdit[] | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/willSaveWaitUntil')
}
