import { Unsubscribable } from 'rxjs'
import { InitializeParams, ServerCapabilities } from '../../protocol'
import { IConnection } from '../server'

/**
 * A proxy for values and methods that exist on the remote client.
 */
export interface Remote extends Partial<Unsubscribable> {
    /**
     * Attach the remote to the given connection.
     *
     * @param connection The connection this remote is operating on.
     */
    attach(connection: IConnection): void

    /**
     * The connection this remote is attached to.
     */
    connection: IConnection

    /**
     * Called to initialize the remote with the given
     * client capabilities
     *
     * @param params the initialization parameters from the client
     */
    initialize(params: InitializeParams): void

    /**
     * Called to fill in the server capabilities this feature implements.
     *
     * @param capabilities The server capabilities to fill.
     */
    fillServerCapabilities(capabilities: ServerCapabilities): void
}
