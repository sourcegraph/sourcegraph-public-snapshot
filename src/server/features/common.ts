import { ClientCapabilities, ServerCapabilities } from '../../protocol'
import { IConnection } from '../server'

/**
 *
 */
export interface Remote {
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
     * @param capabilities The client capabilities
     */
    initialize(capabilities: ClientCapabilities): void

    /**
     * Called to fill in the server capabilities this feature implements.
     *
     * @param capabilities The server capabilities to fill.
     */
    fillServerCapabilities(capabilities: ServerCapabilities): void
}
