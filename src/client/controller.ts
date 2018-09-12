import { BehaviorSubject, Observable, Subject, Subscription, Unsubscribable } from 'rxjs'
import {
    ConfigurationCascade,
    ConfigurationUpdateParams,
    LogMessageParams,
    MessageActionItem,
    ShowInputParams,
    ShowMessageParams,
    ShowMessageRequestParams,
} from '../protocol'
import { Trace } from '../protocol/jsonrpc2/trace'
import { isEqual } from '../util'
import { Client, ClientOptions } from './client'
import { EMPTY_CONTEXT } from './context/context'
import { createObservableEnvironment, EMPTY_ENVIRONMENT, Environment, ObservableEnvironment } from './environment'
import { Extension } from './extension'
import { ContributionFeature } from './features/contribution'
import {
    TextDocumentDefinitionFeature,
    TextDocumentImplementationFeature,
    TextDocumentReferencesFeature,
    TextDocumentTypeDefinitionFeature,
} from './features/location'
import { WindowShowMessageFeature } from './features/message'
import { Registries } from './registries'

/** The minimal unique identifier for a client. */
export interface ClientKey {
    /** The extension ID. */
    id: string
}

/** A client and its unique identifier (key). */
export interface ClientEntry {
    client: Client
    key: ClientKey
}

/** The source of a message (used with LogMessageParams, ShowMessageParams, ShowMessageRequestParams). */
interface MessageSource {
    /** The ID of the extension that produced this log message. */
    extension: string
}

interface PromiseCallback<T> {
    resolve: (p: T | Promise<T>) => void
}

type ShowMessageRequest = ShowMessageRequestParams & MessageSource & PromiseCallback<MessageActionItem | null>

type ShowInputRequest = ShowInputParams & MessageSource & PromiseCallback<string | null>

type ConfigurationUpdate = ConfigurationUpdateParams & MessageSource & PromiseCallback<void>

/**
 * Options for creating the controller.
 *
 * @template X extension type
 * @template C configuration cascade type
 */
export interface ControllerOptions<X extends Extension, C extends ConfigurationCascade> {
    /** Returns additional options to use when creating a client. */
    clientOptions: (
        key: ClientKey,
        options: ClientOptions,
        extension: X
    ) => Pick<
        ClientOptions,
        'createMessageTransports' | 'errorHandler' | 'trace' | 'tracer' | 'experimentalClientCapabilities'
    >

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

    private _clientEntries = new BehaviorSubject<ClientEntry[]>([])

    /** An observable that emits whenever the set of clients managed by this controller changes. */
    public get clientEntries(): Observable<ClientEntry[]> {
        return this._clientEntries
    }

    private subscriptions = new Subscription()

    /** The registries for various providers that expose extension functionality. */
    public readonly registries: Registries<X, C>

    private readonly _logMessages = new Subject<LogMessageParams & MessageSource>()
    private readonly _showMessages = new Subject<ShowMessageParams & MessageSource>()
    private readonly _showMessageRequests = new Subject<ShowMessageRequest>()
    private readonly _showInputs = new Subject<ShowInputRequest>()
    private readonly _configurationUpdates = new Subject<ConfigurationUpdate>()

    /** Log messages from extensions. */
    public readonly logMessages: Observable<LogMessageParams & MessageSource> = this._logMessages

    /** Messages from extensions intended for display to the user. */
    public readonly showMessages: Observable<ShowMessageParams & MessageSource> = this._showMessages

    /** Messages from extensions requesting the user to select an action. */
    public readonly showMessageRequests: Observable<ShowMessageRequest> = this._showMessageRequests

    /** Messages from extensions requesting text input from the user. */
    public readonly showInputs: Observable<ShowInputRequest> = this._showInputs

    /** Configuration updates from extensions. */
    public readonly configurationUpdates: Observable<ConfigurationUpdate> = this._configurationUpdates

    constructor(private options: ControllerOptions<X, C>) {
        this.subscriptions.add(() => {
            for (const c of this._clientEntries.value) {
                c.client.unsubscribe()
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
        const nextClients: ClientEntry[] = []
        const unusedClients: ClientEntry[] = []
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
            unusedClient.client.unsubscribe()
        }
        // Create new clients.
        for (const key of newClients) {
            // Find the extension that this client is for.
            const extension = environment.extensions!.find(x => x.id === key.id)
            if (!extension) {
                throw new Error(`extension not found: ${key.id}`)
            }

            // Construct client.
            const clientOptions: ClientOptions = {
                documentSelector: ['*'],
                createMessageTransports: null as any, // will be overwritten by Object.assign call below
            }
            Object.assign(clientOptions, this.options.clientOptions(key, clientOptions, extension))
            const client = new Client(key.id, clientOptions)

            // Register client features.
            this.registerClientFeatures(client, this.environment.configuration)

            // Activate client.
            client.activate()
            nextClients.push({
                key,
                client,
            })
        }
        this._clientEntries.next(nextClients)
    }

    private registerClientFeatures(client: Client, configuration: Observable<C>): void {
        client.registerFeature(new ContributionFeature(this.registries.contribution))
        client.registerFeature(new TextDocumentDefinitionFeature(client, this.registries.textDocumentDefinition))
        client.registerFeature(
            new TextDocumentImplementationFeature(client, this.registries.textDocumentImplementation)
        )
        client.registerFeature(new TextDocumentReferencesFeature(client, this.registries.textDocumentReferences))
        client.registerFeature(
            new TextDocumentTypeDefinitionFeature(client, this.registries.textDocumentTypeDefinition)
        )
        client.registerFeature(
            new WindowShowMessageFeature(
                client,
                (params: ShowMessageParams) => this._showMessages.next({ ...params, extension: client.id }),
                (params: ShowMessageRequestParams) =>
                    new Promise<MessageActionItem | null>(resolve => {
                        this._showMessageRequests.next({ ...params, extension: client.id, resolve })
                    }),
                (params: ShowInputParams) =>
                    new Promise<string | null>(resolve => {
                        this._showInputs.next({ ...params, extension: client.id, resolve })
                    })
            )
        )
    }

    public readonly environment: ObservableEnvironment<X, C> = createObservableEnvironment<X, C>(this._environment)

    public set trace(value: Trace) {
        for (const client of this._clientEntries.value) {
            client.client.trace = value
        }
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}

function computeClients<X extends Extension>(environment: Pick<Environment<X, any>, 'extensions'>): ClientKey[] {
    const clients: ClientKey[] = []
    if (!environment.extensions) {
        return clients
    }
    for (const x of environment.extensions) {
        clients.push({ id: x.id })
    }
    return clients
}
