import { Range, TextDocumentIdentifier } from 'vscode-languageserver-types'
import { NotificationHandler, RequestHandler } from '../jsonrpc2/handlers'
import { NotificationType, RequestType } from '../jsonrpc2/messages'
import { TextDocumentRegistrationOptions } from './textDocument'

export interface DecorationsClientCapabilities {
    decorations?: DecorationsCapabilityOptions
}

export interface DecorationsCapabilityOptions {
    /**
     * Whether the server supports static decorations (i.e., decorations that the client requests using
     * textDocument/decorations).
     */
    static?: boolean

    /**
     * Whether the server publishes dynamic decorations (i.e., decorations that the server pushes to the client
     * with textDocument/publishDecorations without the client needing to request them).
     */
    dynamic?: boolean
}

export interface DecorationsProviderOptions extends DecorationsCapabilityOptions {}

export interface DecorationsServerCapabilities {
    /** The server's support for decorations. */
    decorationsProvider?: DecorationsProviderOptions | (DecorationsProviderOptions & TextDocumentRegistrationOptions)
}

export interface TextDocumentDecoration {
    range: Range

    isWholeLine?: boolean

    after?: DecorationAttachmentRenderOptions

    background?: string
    backgroundColor?: string
    border?: string
    borderColor?: string
    borderWidth?: string
}

export interface DecorationAttachmentRenderOptions {
    backgroundColor?: string
    color?: string
    contentText?: string
    hoverMessage?: string
    linkURL?: string
}

export interface TextDocumentDecorationsParams {
    textDocument: TextDocumentIdentifier
}

export namespace TextDocumentDecorationsRequest {
    export const type = new RequestType<
        TextDocumentDecorationsParams,
        TextDocumentDecoration[] | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/decorations')
    export type HandlerSignature = RequestHandler<TextDocumentDecorationsParams, TextDocumentDecoration[] | null, void>
}

export interface TextDocumentPublishDecorationsParams {
    textDocument: TextDocumentIdentifier
    decorations: TextDocumentDecoration[] | null
}

export namespace TextDocumentPublishDecorationsNotification {
    export const type = new NotificationType<TextDocumentPublishDecorationsParams, TextDocumentRegistrationOptions>(
        'textDocument/publishDecorations'
    )
    export type HandlerSignature = NotificationHandler<TextDocumentPublishDecorationsParams>
}
