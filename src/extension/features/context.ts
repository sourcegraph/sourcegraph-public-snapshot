import { Context } from '../../environment/context/context'
import { ContextUpdateNotification, ContextUpdateParams } from '../../protocol/context'
import { ExtensionContext, SourcegraphExtensionAPI } from '../api'

/**
 * Creates the Sourcegraph extension API's {@link SourcegraphExtensionAPI#context} value.
 *
 * @param ext The Sourcegraph extension API handle.
 * @return The {@link SourcegraphExtensionAPI#context} value.
 */
export function createExtContext(ext: Pick<SourcegraphExtensionAPI<any>, 'rawConnection'>): ExtensionContext {
    return {
        updateContext: (updates: Context): void => {
            ext.rawConnection.sendNotification(ContextUpdateNotification.type, { updates } as ContextUpdateParams)
        },
    }
}
