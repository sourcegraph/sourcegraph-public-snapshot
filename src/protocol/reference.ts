import { Location, ReferenceContext } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from './textDocument'

/**
 * Parameters for a [ReferencesRequest](#ReferencesRequest).
 */
export interface ReferenceParams extends TextDocumentPositionParams {
    context: ReferenceContext
}

/**
 * A request to resolve project-wide references for the symbol denoted
 * by the given text document position. The request's parameter is of
 * type [ReferenceParams](#ReferenceParams) the response is of type
 * [Location[]](#Location) or a Thenable that resolves to such.
 */
export namespace ReferencesRequest {
    export const type = new RequestType<ReferenceParams, Location[] | null, void, TextDocumentRegistrationOptions>(
        'textDocument/references'
    )
}
