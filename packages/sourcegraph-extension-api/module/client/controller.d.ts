import { Observable, Subscription, Unsubscribable } from 'rxjs';
import { ConfigurationCascade, ConfigurationUpdateParams, LogMessageParams, MessageActionItem, ShowInputParams, ShowMessageParams, ShowMessageRequestParams } from '../protocol';
import { Connection, MessageTransports } from '../protocol/jsonrpc2/connection';
import { Trace } from '../protocol/jsonrpc2/trace';
import { Environment } from './environment';
import { Extension } from './extension';
import { Registries } from './registries';
/** The minimal unique identifier for a client. */
export interface ExtensionConnectionKey {
    /** The extension ID. */
    id: string;
}
/** A connection to an extension and its unique identifier (key). */
export interface ExtensionConnection {
    connection: Promise<Connection>;
    subscription: Subscription;
    key: ExtensionConnectionKey;
}
interface PromiseCallback<T> {
    resolve: (p: T | Promise<T>) => void;
}
declare type ShowMessageRequest = ShowMessageRequestParams & PromiseCallback<MessageActionItem | null>;
declare type ShowInputRequest = ShowInputParams & PromiseCallback<string | null>;
export declare type ConfigurationUpdate = ConfigurationUpdateParams & PromiseCallback<void>;
/**
 * Options for creating the controller.
 *
 * @template X extension type
 * @template C configuration cascade type
 */
export interface ControllerOptions<X extends Extension, C extends ConfigurationCascade> {
    /** Returns additional options to use when creating a client. */
    clientOptions: (key: ExtensionConnectionKey, extension: X) => {
        createMessageTransports: () => MessageTransports | Promise<MessageTransports>;
    };
    /**
     * Called before applying the next environment in Controller#setEnvironment. It should have no side effects.
     */
    environmentFilter?: (nextEnvironment: Environment<X, C>) => Environment<X, C>;
}
/**
 * The controller for the environment.
 *
 * @template X extension type
 * @template C configuration cascade type
 */
export declare class Controller<X extends Extension, C extends ConfigurationCascade> implements Unsubscribable {
    private options;
    private _environment;
    /** The environment. */
    readonly environment: Observable<Environment<X, C>>;
    private _clientEntries;
    /** An observable that emits whenever the set of clients managed by this controller changes. */
    readonly clientEntries: Observable<ExtensionConnection[]>;
    private subscriptions;
    /** The registries for various providers that expose extension functionality. */
    readonly registries: Registries<X, C>;
    private readonly _logMessages;
    private readonly _showMessages;
    private readonly _showMessageRequests;
    private readonly _showInputs;
    private readonly _configurationUpdates;
    /** Log messages from extensions. */
    readonly logMessages: Observable<LogMessageParams>;
    /** Messages from extensions intended for display to the user. */
    readonly showMessages: Observable<ShowMessageParams>;
    /** Messages from extensions requesting the user to select an action. */
    readonly showMessageRequests: Observable<ShowMessageRequest>;
    /** Messages from extensions requesting text input from the user. */
    readonly showInputs: Observable<ShowInputRequest>;
    /** Configuration updates from extensions. */
    readonly configurationUpdates: Observable<ConfigurationUpdate>;
    constructor(options: ControllerOptions<X, C>);
    /**
     * Detect when setEnvironment is called within a setEnvironment call, which probably means there is a bug.
     */
    private inSetEnvironment;
    setEnvironment(nextEnvironment: Environment<X, C>): void;
    private onEnvironmentChange;
    private registerClientFeatures;
    trace: Trace;
    unsubscribe(): void;
}
export {};
