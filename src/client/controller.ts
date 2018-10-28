import { BehaviorSubject, Observable, Subject, Subscription, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'
import { ContextValues } from 'sourcegraph'
import {
    ConfigurationCascade,
    ConfigurationUpdateParams,
    LogMessageParams,
    MessageActionItem,
    ShowInputParams,
    ShowMessageParams,
    ShowMessageRequestParams,
} from '../protocol'
import { Connection, createConnection, MessageTransports } from '../protocol/jsonrpc2/connection'
import { BrowserConsoleTracer, Trace } from '../protocol/jsonrpc2/trace'
import { isEqual } from '../util'
import { ClientCodeEditor } from './api/codeEditor'
import { ClientCommands } from './api/commands'
import { ClientConfiguration } from './api/configuration'
import { ClientContext } from './api/context'
import { ClientDocuments } from './api/documents'
import { ClientLanguageFeatures } from './api/languageFeatures'
import { Search } from './api/search'
import { ClientViews } from './api/views'
import { ClientWindows } from './api/windows'
import { applyContextUpdate, EMPTY_CONTEXT } from './context/context'
import { EMPTY_ENVIRONMENT, Environment } from './environment'
import { Extension } from './extension'
import { Registries } from './registries'

/** The minimal unique identifier for a client. */
export interface ExtensionConnectionKey {
    /** The extension ID. */
    id: string
}

/** A connection to an extension and its unique identifier (key). */
export interface ExtensionConnection {
    connection: Promise<Connection>
    subscription: Subscription
    key: ExtensionConnectionKey
}

interface PromiseCallback<T> {
    resolve: (p: T | Promise<T>) => void
}

type ShowMessageRequest = ShowMessageRequestParams & PromiseCallback<MessageActionItem | null>

type ShowInputRequest = ShowInputParams & PromiseCallback<string | null>

export type ConfigurationUpdate = ConfigurationUpdateParams & PromiseCallback<void>

/**
 * Options for creating the controller.
 *
 * @template X extension type
 * @template C configuration cascade type
 */
export interface ControllerOptions<X extends Extension, C extends ConfigurationCascade> {
    /** Returns additional options to use when creating a client. */
    clientOptions: (
        key: ExtensionConnectionKey,
        extension: X
    ) => { createMessageTransports: () => MessageTransports | Promise<MessageTransports> }

    /**
     * Called before applying the next environment in Controller#setEnvironment. It should have no side effects.
     */
    environmentFilter?: (nextEnvironment: Environment<X, C>) => Environment<X, C>
}

/**
 * The controller for the environment.
 *
 * @template X extension type
 * @template C configuration cascade type
 */
export class Controller<X extends Extension, C extends ConfigurationCascade> implements Unsubscribable {
    private _environment = new BehaviorSubject<Environment<X, C>>(EMPTY_ENVIRONMENT)

    /** The environment. */
    public readonly environment: Observable<Environment<X, C>> = this._environment

    private _clientEntries = new BehaviorSubject<ExtensionConnection[]>([])

    /** An observable that emits whenever the set of clients managed by this controller changes. */
    public get clientEntries(): Observable<ExtensionConnection[]> {
        return this._clientEntries
    }

    private subscriptions = new Subscription()

    /** The registries for various providers that expose extension functionality. */
    public readonly registries: Registries<X, C>

    private readonly _logMessages = new Subject<LogMessageParams>()
    private readonly _showMessages = new Subject<ShowMessageParams>()
    private readonly _showMessageRequests = new Subject<ShowMessageRequest>()
    private readonly _showInputs = new Subject<ShowInputRequest>()
    private readonly _configurationUpdates = new Subject<ConfigurationUpdate>()

    /** Log messages from extensions. */
    public readonly logMessages: Observable<LogMessageParams> = this._logMessages

    /** Messages from extensions intended for display to the user. */
    public readonly showMessages: Observable<ShowMessageParams> = this._showMessages

    /** Messages from extensions requesting the user to select an action. */
    public readonly showMessageRequests: Observable<ShowMessageRequest> = this._showMessageRequests

    /** Messages from extensions requesting text input from the user. */
    public readonly showInputs: Observable<ShowInputRequest> = this._showInputs

    /** Configuration updates from extensions. */
    public readonly configurationUpdates: Observable<ConfigurationUpdate> = this._configurationUpdates

    constructor(private options: ControllerOptions<X, C>) {
        this.subscriptions.add(() => {
            for (const c of this._clientEntries.value) {
                c.subscription.unsubscribe()
            }
        })

        this.registries = new Registries<X, C>(this.environment)
    }

    /**
     * Detect when setEnvironment is called within a setEnvironment call, which probably means there is a bug.
     */
    private inSetEnvironment = false

    public setEnvironment(nextEnvironment: Environment<X, C>): void {
        if (this.inSetEnvironment) {
            throw new Error('setEnvironment may not be called recursively')
        }
        this.inSetEnvironment = true

        if (this.options.environmentFilter) {
            nextEnvironment = this.options.environmentFilter(nextEnvironment)
        }

        // External consumers don't see context, and their setEnvironment args lack context.
        if (nextEnvironment.context === EMPTY_CONTEXT) {
            nextEnvironment = { ...nextEnvironment, context: this._environment.value.context }
        }

        if (isEqual(this._environment.value, nextEnvironment)) {
            this.inSetEnvironment = false
            return // no change
        }

        this._environment.next(nextEnvironment)
        this.onEnvironmentChange()
        this.inSetEnvironment = false
    }

    private onEnvironmentChange(): void {
        const environment = this._environment.value // new environment

        // Diff clients.
        const newClients = computeClients(environment)
        const nextClients: ExtensionConnection[] = []
        const unusedClients: ExtensionConnection[] = []
        for (const oldClient of this._clientEntries.value) {
            const newIndex = newClients.findIndex(newClient => isEqual(oldClient.key, newClient))
            if (newIndex === -1) {
                // Client is no longer needed.
                unusedClients.push(oldClient)
            } else {
                // Client already exists. Settings may have changed, but ConfigurationFeature is responsible for
                // notifying the server of configuration changes.
                newClients.splice(newIndex, 1)
                nextClients.push(oldClient)
            }
        }
        // Remove clients that are no longer in use.
        for (const unusedClient of unusedClients) {
            unusedClient.subscription.unsubscribe()
        }
        // Create new clients.
        for (const key of newClients) {
            // Find the extension that this client is for.
            const extension = environment.extensions!.find(x => x.id === key.id)
            if (!extension) {
                throw new Error(`extension not found: ${key.id}`)
            }

            // Construct client.
            const clientOptions = this.options.clientOptions(key, extension)
            const subscription = new Subscription()
            const extensionConnection: ExtensionConnection = {
                key,
                subscription,
                connection: Promise.resolve(clientOptions.createMessageTransports()).then(transports => {
                    const connection = createConnection(transports)
                    subscription.add(connection)
                    connection.listen()
                    connection.onRequest('ping', () => 'pong')
                    this.registerClientFeatures(
                        connection,
                        subscription,
                        this.environment.pipe(
                            map(({ configuration }) => configuration),
                            distinctUntilChanged()
                        )
                    )
                    return connection
                }),
            }

            nextClients.push(extensionConnection)
        }
        this._clientEntries.next(nextClients)
    }

    private registerClientFeatures(client: Connection, subscription: Subscription, configuration: Observable<C>): void {
        subscription.add(
            new ClientConfiguration(
                client,
                configuration,
                (params: ConfigurationUpdateParams) =>
                    new Promise<void>(resolve => this._configurationUpdates.next({ ...params, resolve }))
            )
        )
        subscription.add(
            new ClientContext(client, (updates: ContextValues) =>
                // Set environment manually, not via Controller#setEnvironment, to avoid recursive setEnvironment calls
                // (when this callback is called during setEnvironment's teardown of unused clients).
                this._environment.next({
                    ...this._environment.value,
                    context: applyContextUpdate(this._environment.value.context, updates),
                })
            )
        )
        subscription.add(
            new ClientWindows(
                client,
                this.environment.pipe(
                    map(({ visibleTextDocuments }) => visibleTextDocuments),
                    distinctUntilChanged()
                ),
                (params: ShowMessageParams) => this._showMessages.next({ ...params }),
                (params: ShowMessageRequestParams) =>
                    new Promise<MessageActionItem | null>(resolve => {
                        this._showMessageRequests.next({ ...params, resolve })
                    }),
                (params: ShowInputParams) =>
                    new Promise<string | null>(resolve => {
                        this._showInputs.next({ ...params, resolve })
                    })
            )
        )
        subscription.add(new ClientViews(client, this.registries.views))
        subscription.add(new ClientCodeEditor(client, this.registries.textDocumentDecoration))
        subscription.add(
            new ClientDocuments(
                client,
                this.environment.pipe(
                    map(({ visibleTextDocuments }) => visibleTextDocuments),
                    distinctUntilChanged()
                )
            )
        )
        subscription.add(
            new ClientLanguageFeatures(
                client,
                this.registries.textDocumentHover,
                this.registries.textDocumentDefinition,
                this.registries.textDocumentTypeDefinition,
                this.registries.textDocumentImplementation,
                this.registries.textDocumentReferences
            )
        )
        subscription.add(new Search(client, this.registries.queryTransformer))
        subscription.add(new ClientCommands(client, this.registries.commands))
    }

    public set trace(value: Trace) {
        for (const client of this._clientEntries.value) {
            client.connection
                .then(connection => {
                    connection.trace(value, new BrowserConsoleTracer(client.key.id))
                })
                .catch(() => void 0)
        }
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}

function computeClients<X extends Extension>(
    environment: Pick<Environment<X, any>, 'extensions'>
): ExtensionConnectionKey[] {
    const clients: ExtensionConnectionKey[] = []
    if (!environment.extensions) {
        return clients
    }
    for (const x of environment.extensions) {
        clients.push({ id: x.id })
    }
    return clients
}
