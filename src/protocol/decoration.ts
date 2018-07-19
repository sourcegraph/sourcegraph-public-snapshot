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
     * textDocument/decoration).
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

/**
 * A text document decoration changes the appearance of a range in the document and/or adds other content to it.
 */
export interface TextDocumentDecoration {
    /** The range that the decoration applies to. */
    range: Range

    /**
     * If true, the decoration applies to all lines in the range (inclusive), even if not all characters on the
     * line are included.
     */
    isWholeLine?: boolean

    /** Content to display after the range. */
    after?: DecorationAttachmentRenderOptions

    /** The CSS background-color property value for the line. */
    backgroundColor?: string

    /** The CSS border property value for the line. */
    border?: string

    /** The CSS border-color property value for the line. */
    borderColor?: string

    /** The CSS border-width property value for the line. */
    borderWidth?: string
}

/** A decoration attachment adds content after a [decoration](#TextDocumentDecoration). */
export interface DecorationAttachmentRenderOptions {
    /** The CSS background-color property value for the attachment. */
    backgroundColor?: string

    /** The CSS color property value for the attachment. */
    color?: string

    /** Text to display in the attachment. */
    contentText?: string

    /** Tooltip text to display when hovering over the attachment. */
    hoverMessage?: string

    /** If set, the attachment becomes a link with this destination URL. */
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
