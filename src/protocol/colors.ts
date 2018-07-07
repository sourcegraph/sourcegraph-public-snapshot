import { Range, TextDocumentIdentifier, TextEdit } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'
import { StaticRegistrationOptions } from './registration'
import { TextDocumentRegistrationOptions } from './textDocument'

// ---- Server capability ----

export interface ColorClientCapabilities {
    /**
     * The text document client capabilities
     */
    textDocument?: {
        /**
         * Capabilities specific to the colorProvider
         */
        colorProvider?: {
            /**
             * Whether implementation supports dynamic registration. If this is set to `true`
             * the client supports the new `(ColorProviderOptions & TextDocumentRegistrationOptions & StaticRegistrationOptions)`
             * return value for the corresponding server capability as well.
             */
            dynamicRegistration?: boolean
        }
    }
}

export interface ColorProviderOptions {}

export interface ColorServerCapabilities {
    /**
     * The server provides color provider support.
     */
    colorProvider?:
        | boolean
        | ColorProviderOptions
        | (ColorProviderOptions & TextDocumentRegistrationOptions & StaticRegistrationOptions)
}

/**
 * Parameters for a [DocumentColorRequest](#DocumentColorRequest).
 */
export interface DocumentColorParams {
    /**
     * The text document.
     */
    textDocument: TextDocumentIdentifier
}

/**
 * A request to list all color symbols found in a given text document. The request's
 * parameter is of type [DocumentColorParams](#DocumentColorParams) the
 * response is of type [ColorInformation[]](#ColorInformation) or a Thenable
 * that resolves to such.
 */
export namespace DocumentColorRequest {
    export const type = new RequestType<DocumentColorParams, ColorInformation[], void, TextDocumentRegistrationOptions>(
        'textDocument/documentColor'
    )
}

/**
 * Parameters for a [ColorPresentationRequest](#ColorPresentationRequest).
 */
export interface ColorPresentationParams {
    /**
     * The text document.
     */
    textDocument: TextDocumentIdentifier

    /**
     * The color to request presentations for.
     */
    color: Color

    /**
     * The range where the color would be inserted. Serves as a context.
     */
    range: Range
}

/**
 * A request to list all presentation for a color. The request's
 * parameter is of type [ColorPresentationParams](#ColorPresentationParams) the
 * response is of type [ColorInformation[]](#ColorInformation) or a Thenable
 * that resolves to such.
 */
export namespace ColorPresentationRequest {
    export const type = new RequestType<
        ColorPresentationParams,
        ColorPresentation[],
        void,
        TextDocumentRegistrationOptions
    >('textDocument/colorPresentation')
}

/**
 * Represents a color in RGBA space.
 */
export interface Color {
    /**
     * The red component of this color in the range [0-1].
     */
    readonly red: number

    /**
     * The green component of this color in the range [0-1].
     */
    readonly green: number

    /**
     * The blue component of this color in the range [0-1].
     */
    readonly blue: number

    /**
     * The alpha component of this color in the range [0-1].
     */
    readonly alpha: number
}

/**
 * Represents a color range from a document.
 */
export interface ColorInformation {
    /**
     * The range in the document where this color appers.
     */
    range: Range

    /**
     * The actual color value for this color range.
     */
    color: Color
}

export interface ColorPresentation {
    /**
     * The label of this color presentation. It will be shown on the color
     * picker header. By default this is also the text that is inserted when selecting
     * this color presentation.
     */
    label: string
    /**
     * An [edit](#TextEdit) which is applied to a document when selecting
     * this presentation for the color.  When `falsy` the [label](#ColorPresentation.label)
     * is used.
     */
    textEdit?: TextEdit
    /**
     * An optional array of additional [text edits](#TextEdit) that are applied when
     * selecting this color presentation. Edits must not overlap with the main [edit](#ColorPresentation.textEdit) nor with themselves.
     */
    additionalTextEdits?: TextEdit[]
}
