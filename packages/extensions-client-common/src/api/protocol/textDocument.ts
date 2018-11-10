import { DocumentSelector } from 'sourcegraph'
import { TextDocumentIdentifier } from '../client/types/textDocument'
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
 * General text document registration options.
 */
export interface TextDocumentRegistrationOptions {
    /**
     * A document selector to identify the scope of the registration. If set to null
     * the document selector provided on the client side will be used.
     */
    documentSelector: DocumentSelector | null
}
