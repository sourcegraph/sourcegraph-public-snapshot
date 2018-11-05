import { Subscription } from 'rxjs'
import { ContextValues } from 'sourcegraph'
import { handleRequests } from '../../common/proxy'
import { Connection } from '../../protocol/jsonrpc2/connection'

/** @internal */
export interface ClientContextAPI {
    $acceptContextUpdates(updates: ContextValues): void
}

/** @internal */
export class ClientContext implements ClientContextAPI {
    private subscriptions = new Subscription()

    /**
     * Context keys set by this server. To ensure that context values are cleaned up, all context properties that
     * the server set are cleared upon deinitialization. This errs on the side of clearing too much (if another
     * server set the same context keys after this server, then those keys would also be cleared when this server's
     * client deinitializes).
     */
    private keys = new Set<string>()

    constructor(connection: Connection, private updateContext: (updates: ContextValues) => void) {
        handleRequests(connection, 'context', this)
    }

    public $acceptContextUpdates(updates: ContextValues): void {
        for (const key of Object.keys(updates)) {
            this.keys.add(key)
        }
        this.updateContext(updates)
    }

    public unsubscribe(): void {
        /**
         * Clear all context properties whose keys were ever set by the server. See {@link ClientContext#keys}.
         */
        const updates: ContextValues = {}
        for (const key of this.keys) {
            updates[key] = null
        }
        this.keys.clear()
        this.updateContext(updates)

        this.subscriptions.unsubscribe()
    }
}
