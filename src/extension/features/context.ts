import { Context } from '../../environment/context/context'
import { ContextUpdateNotification, ContextUpdateParams } from '../../protocol/context'
import { CXP, ExtensionContext } from '../api'

/**
 * Creates the CXP extension API's {@link CXP#context} value.
 *
 * @param ext The CXP extension API handle.
 * @return The {@link CXP#context} value.
 */
export function createExtContext(ext: Pick<CXP<any>, 'rawConnection'>): ExtensionContext {
    return {
        updateContext: (updates: Context): void => {
            ext.rawConnection.sendNotification(ContextUpdateNotification.type, { updates } as ContextUpdateParams)
        },
    }
}
