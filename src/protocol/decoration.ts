import { Range, TextDocumentIdentifier } from 'vscode-languageserver-types'
import { NotificationHandler, RequestHandler } from '../jsonrpc2/handlers'
import { NotificationType, RequestType } from '../jsonrpc2/messages'
import { TextDocumentRegistrationOptions } from './textDocument'

export interface DecorationClientCapabilities {
    decoration?: DecorationCapabilityOptions
}

export interface DecorationCapabilityOptions {
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

export interface DecorationProviderOptions extends DecorationCapabilityOptions {}

export interface DecorationServerCapabilities {
    /** The server's support for decoration. */
    decorationProvider?: DecorationProviderOptions | (DecorationProviderOptions & TextDocumentRegistrationOptions)
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

export interface TextDocumentDecorationParams {
    textDocument: TextDocumentIdentifier
}

export namespace TextDocumentDecorationRequest {
    export const type = new RequestType<
        TextDocumentDecorationParams,
        TextDocumentDecoration[] | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/decoration')
    export type HandlerSignature = RequestHandler<TextDocumentDecorationParams, TextDocumentDecoration[] | null, void>
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
