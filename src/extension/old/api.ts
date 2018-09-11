import { Subscription } from 'rxjs'
import { InitializeParams, Settings } from '../../protocol'
import { Connection } from '../../protocol/jsonrpc2/connection'

/**
 * The Sourcegraph extension API, which extensions use to interact with the client.
 *
 * @template C the extension's settings
 */
export interface SourcegraphExtensionAPI<C = Settings> {
    /**
     * The params passed by the client in the `initialize` request.
     */
    initializeParams: InitializeParams

    /**
     * The configuration settings from the client.
     */
    configuration: Configuration<C>

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
 * Configuration settings for a specific resource (such as a file, directory, or repository) or subject (such as a
 * user or organization, depending on the client).
 *
 * It may be merged from the following sources of settings, in order:
 *
 * - Default settings
 * - Global settings
 * - Organization settings (for all organizations the user is a member of)
 * - User settings
 * - Client settings
 * - Repository settings
 * - Directory settings
 *
 * @template C configuration type
 */
export interface Configuration<C> extends Observable<C> {
    /**
     * Returns a value from the configuration.
     *
     * @template K Valid keys on the configuration object.
     * @param key The name of the configuration property to get.
     * @return The configuration value, or undefined.
     */
    get<K extends keyof C>(key: K): C[K] | undefined

    /**
     * Observes changes to the configuration values for the given keys.
     *
     * @template K Valid keys on the configuration object.
     * @param keys The names of the configuration properties to observe.
     * @return An observable that emits when any of the keys' values change (using deep comparison).
     */
    watch<K extends keyof C>(...keys: K[]): Observable<Pick<C, K>>

    /**
     * Updates the configuration value for the given key. The updated configuration value is sent to the client for
     * persistence.
     *
     * @template K Valid keys on the configuration object.
     * @param key The name of the configuration property to update.
     * @param value The new value, or undefined to remove it.
     * @return A promise that resolves when the client acknowledges the update.
     */
    update<K extends keyof C>(key: K, value: C[K] | undefined): Promise<void>

    // TODO: Future plans:
    //
    // - add a way to read configuration from a specific scope (aka subject, but "scope" is probably a better word)
    // - describe how configuration defaults are supported
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
