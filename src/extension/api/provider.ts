import { Unsubscribable } from 'sourcegraph'
import * as sourcegraph from 'sourcegraph'
import {
    HoverRequest,
    RegistrationParams,
    RegistrationRequest,
    TextDocumentPositionParams,
    TextDocumentRegistrationOptions,
    UnregistrationParams,
    UnregistrationRequest,
} from '../../protocol'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { idSequence } from '../../util'
import { Position } from '../types/position'

/**
 * All `registerXyzProvider` functions in the Sourcegraph extension API.
 */
interface RegisterProviderFunctions extends Pick<typeof sourcegraph, 'registerHoverProvider'> {}

/**
 * Create all `registerXyzProvider` functions in the Sourcegraph extension API.
 */
export function createRegisterProviderFunctions(connection: Connection): RegisterProviderFunctions {
    return {
        registerHoverProvider: (selector, provider) => {
            connection.onRequest(HoverRequest.type, (params: TextDocumentPositionParams) =>
                provider.provideHover(
                    params.textDocument as sourcegraph.TextDocument,
                    new Position(params.position.character, params.position.character)
                )
            )
            return registerProvider<TextDocumentRegistrationOptions>(connection, HoverRequest.type, {
                documentSelector: selector,
                extensionID: '', // TODO(sqs): use provider ID
            })
        },
    }
}

/**
 * Registers a provider implemented by the extension with the client.
 *
 * @return An {@link Unsubscribable} that unregisters the provider.
 */
function registerProvider<RO>(connection: Connection, method: string, registerOptions: RO): Unsubscribable {
    const id = idSequence()
    // TODO(sqs): handle errors in sendRequest calls
    connection
        .sendRequest(RegistrationRequest.type, {
            registrations: [{ id, method, registerOptions }],
        } as RegistrationParams)
        .catch(err => console.error(err))
    return {
        unsubscribe: () =>
            connection
                .sendRequest(UnregistrationRequest.type, {
                    unregisterations: [{ id, method }],
                } as UnregistrationParams)
                .catch(err => console.error(err)),
    }
}
