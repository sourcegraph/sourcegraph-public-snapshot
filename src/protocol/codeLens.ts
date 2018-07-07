import { CodeLens, TextDocumentIdentifier } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'
import { TextDocumentRegistrationOptions } from './textDocument'

/**
 * Code Lens options.
 */
export interface CodeLensOptions {
    /**
     * Code lens has a resolve provider as well.
     */
    resolveProvider?: boolean
}

/**
 * Params for the Code Lens request.
 */
export interface CodeLensParams {
    /**
     * The document to request code lens for.
     */
    textDocument: TextDocumentIdentifier
}

/**
 * Code Lens registration options.
 */
export interface CodeLensRegistrationOptions extends TextDocumentRegistrationOptions, CodeLensOptions {}

/**
 * A request to provide code lens for the given text document.
 */
export namespace CodeLensRequest {
    export const type = new RequestType<CodeLensParams, CodeLens[] | null, void, CodeLensRegistrationOptions>(
        'textDocument/codeLens'
    )
}

/**
 * A request to resolve a command for a given code lens.
 */
export namespace CodeLensResolveRequest {
    export const type = new RequestType<CodeLens, CodeLens, void, void>('codeLens/resolve')
}
