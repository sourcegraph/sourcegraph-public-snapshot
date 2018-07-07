import { SignatureHelp } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from './textDocument'

/**
 * Signature help options.
 */
export interface SignatureHelpOptions {
    /**
     * The characters that trigger signature help
     * automatically.
     */
    triggerCharacters?: string[]
}

/**
 * Signature help registration options.
 */
export interface SignatureHelpRegistrationOptions extends TextDocumentRegistrationOptions, SignatureHelpOptions {}

export namespace SignatureHelpRequest {
    export const type = new RequestType<
        TextDocumentPositionParams,
        SignatureHelp | null,
        void,
        SignatureHelpRegistrationOptions
    >('textDocument/signatureHelp')
}
