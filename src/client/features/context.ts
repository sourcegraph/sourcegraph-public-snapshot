import { ContextUpdateNotification, ContextUpdateParams } from '../../protocol/context'
import { Client } from '../client'
import { StaticFeature } from './common'

/**
 * Support for context set by the server (context/update notifications from the server).
 */
export class ContextFeature implements StaticFeature {
    /**
     * Context keys set by this server. To ensure that context values are cleaned up, all context properties that
     * the server set are cleared upon deinitialization. This errs on the side of clearing too much (if another
     * server set the same context keys after this server, then those keys would also be cleared when this server's
     * client deinitializes).
     */
    private keys = new Set<string>()

    constructor(private client: Client, private setContext: (params: ContextUpdateParams) => void) {}

    public readonly messages = ContextUpdateNotification.type

    public initialize(): void {
        this.client.onNotification(ContextUpdateNotification.type, params => {
            for (const key of Object.keys(params.updates)) {
                this.keys.add(key)
            }
            this.setContext(params)
        })
    }

    public deinitialize(): void {
        // Clear all context properties whose keys were ever set by the server. See ContextFeature#keys.
        const params: ContextUpdateParams = { updates: {} }
        for (const [key] of this.keys.entries()) {
            params.updates[key] = null
        }
        this.keys.clear()
        this.setContext(params)
    }
}
