import { Subscription } from 'rxjs'
import { InitializeParams } from '../../protocol'
import { Connection } from '../../protocol/jsonrpc2/connection'

/**
 * The Sourcegraph extension API, which extensions use to interact with the client.
 *
 * @template C the extension's settings
 */
export interface SourcegraphExtensionAPI<C = any> {
    /**
     * The params passed by the client in the `initialize` request.
     */
    initializeParams: InitializeParams

    /**
     * Command registration and execution.
     */
    commands: Commands

    /**
     * The underlying connection to the Sourcegraph extension client.
     * @internal
     */
    readonly rawConnection: Connection

    /**
     * Immediately stops the extension and closes the connection to the client.
     */
    close(): void
}

/**
 * A stream of values that can be transformed (with {@link Observable#pipe}) and subscribed to (with
 * {@link Observable#subscribe}).
 *
 * This is a subset of the {@link module:rxjs.Observable} interface, for simplicity and compatibility with future
 * Observable standards.
 *
 * @template T The type of the values emitted by the {@link Observable}.
 */
export interface Observable<T> {
    /**
     * Registers callbacks that are called each time a certain event occurs in the stream of values.
     *
     * @param next Called when a new value is emitted in the stream.
     * @param error Called when an error occurs (which also causes the observable to be closed).
     * @param complete Called when the stream of values ends.
     * @return A subscription that frees resources used by the subscription upon unsubscription.
     */
    subscribe(next?: (value: T) => void, error?: (error: any) => void, complete?: () => void): Subscription

    /**
     * Returns the underlying Observable value, for compatibility with other Observable implementations (such as
     * RxJS).
     *
     * @internal
     */
    [Symbol.observable]?(): any
}

/**
 * Command registration and execution.
 */
export interface Commands {
    /**
     * Registers a command with the given identifier. The command can be invoked by this extension's contributions
     * (e.g., a contributed action that adds a toolbar item to invoke this command).
     *
     * @param command The unique identifier for the command.
     * @param run The function to invoke for this command.
     * @return A subscription that unregisters this command upon unsubscription.
     */
    register(command: string, run: (...args: any[]) => any): Subscription
}
