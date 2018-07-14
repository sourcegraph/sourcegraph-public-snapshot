import { Diagnostic } from 'vscode-languageserver-types'
import { NotificationType } from '../jsonrpc2/messages'

/**
 * Diagnostics notification are sent from the server to the client to signal
 * results of validation runs.
 */
export namespace PublishDiagnosticsNotification {
    export const type = new NotificationType<PublishDiagnosticsParams, void>('textDocument/publishDiagnostics')
}

/**
 * The publish diagnostic notification's parameters.
 */
export interface PublishDiagnosticsParams {
    /**
     * The URI for which diagnostic information is reported.
     */
    uri: string

    /**
     * An array of diagnostic information items.
     */
    diagnostics: Diagnostic[]
}
