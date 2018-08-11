import { BehaviorSubject, Observable, Subject, Subscription, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, filter, map } from 'rxjs/operators'
import { Client, ClientOptions } from '../client/client'
import { ExecuteCommandFeature } from '../client/features/command'
import {
    ConfigurationChangeNotificationFeature,
    ConfigurationFeature,
    ConfigurationUpdateFeature,
} from '../client/features/configuration'
import { ContributionFeature } from '../client/features/contribution'
import { TextDocumentDecorationFeature } from '../client/features/decoration'
import { TextDocumentHoverFeature } from '../client/features/hover'
import {
    TextDocumentDefinitionFeature,
    TextDocumentImplementationFeature,
    TextDocumentReferencesFeature,
    TextDocumentTypeDefinitionFeature,
} from '../client/features/location'
import { WindowLogMessageFeature } from '../client/features/logMessage'
import { WindowShowMessageFeature } from '../client/features/message'
import { TextDocumentDidCloseFeature, TextDocumentDidOpenFeature } from '../client/features/textDocument'
import { Trace } from '../jsonrpc2/trace'
import {
    ConfigurationUpdateParams,
    InitializeParams,
    LogMessageParams,
    MessageActionItem,
    ShowInputParams,
    ShowMessageParams,
    ShowMessageRequestParams,
} from '../protocol'
import { isEqual } from '../util'
import { createObservableEnvironment, EMPTY_ENVIRONMENT, Environment, ObservableEnvironment } from './environment'
import { Extension, ExtensionSettings } from './extension'
import { Registries } from './registries'

/** The minimal unique identifier for a client. */
export interface ClientKey extends Pick<InitializeParams, 'root' | 'initializationOptions'> {
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

/** Options for creating the controller. */
export interface ControllerOptions<X extends Extension> {
    /** Returns additional options to use when creating a client. */
    clientOptions: (
        key: ClientKey,
        options: ClientOptions,
        extension: X
    ) => Pick<
        ClientOptions,
        'middleware' | 'createMessageTransports' | 'errorHandler' | 'initializationFailedHandler' | 'trace' | 'tracer'
    >

    /**
     * Called before applying the next environment in Controller#setEnvironment. It should have no side effects.
     */
    environmentFilter?: (nextEnvironment: Environment<X>) => Environment<X>
}

/**
 * The controller for the environment.
 */
export class Controller<X extends Extension = Extension> implements Unsubscribable {
    private _environment = new BehaviorSubject<Environment<X>>(EMPTY_ENVIRONMENT)

    private _clientEntries = new BehaviorSubject<ClientEntry[]>([])

    /** An observable that emits whenever the set of clients managed by this controller changes. */
    public get clientEntries(): Observable<ClientEntry[]> {
        return this._clientEntries
    }

    private subscriptions = new Subscription()

    /** The registries for various providers that expose extension functionality. */
    public readonly registries = new Registries()

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

    constructor(private options: ControllerOptions<X>) {
        this.subscriptions.add(() => {
            for (const c of this._clientEntries.value) {
                c.client.unsubscribe()
            }
        })
    }

    public setEnvironment(nextEnvironment: Environment<X>): void {
        if (this.options.environmentFilter) {
            nextEnvironment = this.options.environmentFilter(nextEnvironment)
        }

        if (isEqual(this._environment.value, nextEnvironment)) {
            return // no change
        }

        this._environment.next(nextEnvironment)
        this.onEnvironmentChange()
    }

    private onEnvironmentChange(): void {
        const environment = this._environment.value // new environment

        // Diff clients.
        const newClients = computeClients(environment)
        const nextClients: ClientEntry[] = []
        const unusedClients: ClientEntry[] = []
        for (const oldClient of this._clientEntries.value) {
            const newIndex = newClients.findIndex(({ key }) => isEqual(oldClient.key as ClientKey, key as ClientKey))
            if (newIndex === -1) {
                // Client is no longer needed.
                unusedClients.push(oldClient)
            } else {
                // Client already exists. Settings may have changed, but ConfigurationFeature is responsible for
                // notifying the server of settings changes.
                newClients.splice(newIndex, 1)
                nextClients.push(oldClient)
            }
        }
        // Remove clients that are no longer in use.
        for (const unusedClient of unusedClients) {
            unusedClient.client.unsubscribe()
        }
        // Create new clients.
        for (const { key } of newClients) {
            // Find the extension that this client is for.
            const extension = environment.extensions!.find(x => x.id === key.id)
            if (!extension) {
                throw new Error(`extension not found: ${key.id}`)
            }

            // Construct client.
            const clientOptions: ClientOptions = {
                root: key.root,
                initializationOptions: { ...key.initializationOptions }, // key is immutable so we can diff it
                documentSelector: ['*'],
                createMessageTransports: null as any, // will be overwritten by Object.assign call below
            }
            Object.assign(clientOptions, this.options.clientOptions(key, clientOptions, extension))
            const client = new Client(key.id, key.id, clientOptions)

            // Register client features.
            const settings = this._environment.pipe(
                map(({ extensions }) => (extensions ? extensions.find(x => x.id === key.id) : null)),
                filter((x): x is X => !!x),
                map(x => x.settings),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
            this.registerClientFeatures(client, settings)

            // Activate client.
            client.activate()
            nextClients.push({
                key,
                client,
            })
        }
        this._clientEntries.next(nextClients)
    }

    private registerClientFeatures(client: Client, settings: Observable<ExtensionSettings>): void {
        client.registerFeature(new ConfigurationChangeNotificationFeature(client, settings))
        client.registerFeature(new ConfigurationFeature(client, settings))
        client.registerFeature(
            new ConfigurationUpdateFeature(
                client,
                (params: ConfigurationUpdateParams) =>
                    new Promise<void>(resolve =>
                        this._configurationUpdates.next({ ...params, extension: client.id, resolve })
                    )
            )
        )
        client.registerFeature(new ContributionFeature(this.registries.contribution))
        client.registerFeature(new ExecuteCommandFeature(client, this.registries.commands))
        client.registerFeature(new TextDocumentDidOpenFeature(client, this.environment))
        client.registerFeature(new TextDocumentDidCloseFeature(client, this.environment))
        client.registerFeature(new TextDocumentDefinitionFeature(client, this.registries.textDocumentDefinition))
        client.registerFeature(
            new TextDocumentImplementationFeature(client, this.registries.textDocumentImplementation)
        )
        client.registerFeature(new TextDocumentReferencesFeature(client, this.registries.textDocumentReferences))
        client.registerFeature(
            new TextDocumentTypeDefinitionFeature(client, this.registries.textDocumentTypeDefinition)
        )
        client.registerFeature(new TextDocumentHoverFeature(client, this.registries.textDocumentHover))
        client.registerFeature(new TextDocumentDecorationFeature(client, this.registries.textDocumentDecoration))
        client.registerFeature(
            new WindowLogMessageFeature(client, (params: LogMessageParams) =>
                this._logMessages.next({ ...params, extension: client.id })
            )
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

    public readonly environment: ObservableEnvironment<X> = createObservableEnvironment<X>(this._environment)

    public set trace(value: Trace) {
        for (const client of this._clientEntries.value) {
            client.client.trace = value
        }
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}

interface ClientInit {
    key: ClientKey
    settings: ExtensionSettings
}

function computeClients<X extends Extension>(environment: Environment<X>): ClientInit[] {
    const clients: ClientInit[] = []
    if (!environment.extensions) {
        return clients
    }
    for (const x of environment.extensions) {
        clients.push({
            key: {
                id: x.id,
                root: environment.root,
                initializationOptions: {
                    // TODO(sqs): Add a type for InitializationOptions sent to the Sourcegraph CXP proxy.
                    session: 'cxp', // the special 'cxp' value makes each connection an isolated session
                    mode: x.id,
                    // Note: settings are omitted here because they do not form part of the client key (because
                    // merely changing settings does not require a new client to be created for the new settings).
                    // They are filled in by ConfigurationFeature.
                },
            },
            settings: x.settings,
        })
    }
    return clients
}
