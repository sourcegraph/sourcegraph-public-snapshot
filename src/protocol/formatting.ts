import { FormattingOptions, Position, Range, TextDocumentIdentifier, TextEdit } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'
import { TextDocumentRegistrationOptions } from './textDocument'

/**
 * Format document on type options
 */
export interface DocumentOnTypeFormattingOptions {
    /**
     * A character on which formatting should be triggered, like `}`.
     */
    firstTriggerCharacter: string

    /**
     * More trigger characters.
     */
    moreTriggerCharacter?: string[]
}

export interface DocumentFormattingParams {
    /**
     * The document to format.
     */
    textDocument: TextDocumentIdentifier

    /**
     * The format options
     */
    options: FormattingOptions
}

/**
 * A request to to format a whole document.
 */
export namespace DocumentFormattingRequest {
    export const type = new RequestType<
        DocumentFormattingParams,
        TextEdit[] | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/formatting')
}

export interface DocumentRangeFormattingParams {
    /**
     * The document to format.
     */
    textDocument: TextDocumentIdentifier

    /**
     * The range to format
     */
    range: Range

    /**
     * The format options
     */
    options: FormattingOptions
}

/**
 * A request to to format a range in a document.
 */
export namespace DocumentRangeFormattingRequest {
    export const type = new RequestType<
        DocumentRangeFormattingParams,
        TextEdit[] | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/rangeFormatting')
}

export interface DocumentOnTypeFormattingParams {
    /**
     * The document to format.
     */
    textDocument: TextDocumentIdentifier

    /**
     * The position at which this request was send.
     */
    position: Position

    /**
     * The character that has been typed.
     */
    ch: string

    /**
     * The format options.
     */
    options: FormattingOptions
}

/**
 * Format document on type options
 */
export interface DocumentOnTypeFormattingRegistrationOptions
    extends TextDocumentRegistrationOptions,
        DocumentOnTypeFormattingOptions {}

/**
 * A request to format a document on type.
 */
export namespace DocumentOnTypeFormattingRequest {
    export const type = new RequestType<
        DocumentOnTypeFormattingParams,
        TextEdit[] | null,
        void,
        DocumentOnTypeFormattingRegistrationOptions
    >('textDocument/onTypeFormatting')
}
