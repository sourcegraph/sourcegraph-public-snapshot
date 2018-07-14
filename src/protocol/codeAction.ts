import { CodeAction, CodeActionContext, Command, Range, TextDocumentIdentifier } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'
import { TextDocumentRegistrationOptions } from './textDocument'

/**
 * Params for the CodeActionRequest
 */
export interface CodeActionParams {
    /**
     * The document in which the command was invoked.
     */
    textDocument: TextDocumentIdentifier

    /**
     * The range for which the command was invoked.
     */
    range: Range

    /**
     * Context carrying additional information.
     */
    context: CodeActionContext
}

/**
 * A request to provide commands for the given text document and range.
 */
export namespace CodeActionRequest {
    export const type = new RequestType<
        CodeActionParams,
        (Command | CodeAction)[] | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/codeAction')
}
