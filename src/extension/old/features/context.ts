import { Context } from '../../../client/context/context'
import { ContextUpdateNotification, ContextUpdateParams } from '../../../protocol/context'
import { MessageConnection } from '../../../protocol/jsonrpc2/connection'
import { ExtensionContext } from '../api'

/**
 * Creates the Sourcegraph extension API's {@link SourcegraphExtensionAPI#context} value.
 *
 * @param rawConnection The connection to the Sourcegraph API client.
 * @return The {@link SourcegraphExtensionAPI#context} value.
 */
export function createExtContext(rawConnection: MessageConnection): ExtensionContext {
    return {
        updateContext: (updates: Context): void => {
            rawConnection.sendNotification(ContextUpdateNotification.type, { updates } as ContextUpdateParams)
        },
    }
}
