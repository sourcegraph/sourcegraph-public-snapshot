import { Position } from 'vscode-languageserver-types'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'

/** Useful test fixtures. */
export const FIXTURE = {
    TextDocumentPositionParams: {
        position: Position.create(1, 2),
        textDocument: { uri: 'file:///f' },
    } as TextDocumentPositionParams,

    PartialEntry: {
        registrationOptions: {
            documentSelector: ['*'],
        } as TextDocumentRegistrationOptions,
    },
}
