import { BehaviorSubject, Observable, Subject, Subscription, SubscriptionLike, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, filter, map } from 'rxjs/operators'
import { Client, ClientOptions } from '../client/client'
import { ErrorHandler, InitializationFailedHandler } from '../client/errorHandler'
import { ExecuteCommandFeature } from '../client/features/command'
import {
    ConfigurationChangeNotificationFeature,
    ConfigurationFeature,
    ConfigurationUpdateFeature,
} from '../client/features/configuration'
import {
    TextDocumentDynamicDecorationFeature,
    TextDocumentStaticDecorationFeature,
} from '../client/features/decoration'
import { TextDocumentHoverFeature } from '../client/features/hover'
import {
    TextDocumentDefinitionFeature,
    TextDocumentImplementationFeature,
    TextDocumentReferencesFeature,
    TextDocumentTypeDefinitionFeature,
} from '../client/features/location'
import { WindowLogMessageFeature } from '../client/features/logMessage'
import { WindowShowMessageFeature } from '../client/features/message'
import { TextDocumentDidOpenFeature } from '../client/features/textDocument'
import { MessageTransports } from '../jsonrpc2/connection'
import { Trace } from '../jsonrpc2/trace'
import {
    ConfigurationUpdateParams,
    InitializeParams,
    LogMessageParams,
    MessageActionItem,
    ShowMessageParams,
    ShowMessageRequestParams,
} from '../protocol'
import { isEqual } from '../util'
import { createObservableEnvironment, EMPTY_ENVIRONMENT, Environment, ObservableEnvironment } from './environment'
import { Extension, ExtensionSettings } from './extension'
import { Registries } from './registries'

interface ClientKey extends Pick<InitializeParams, 'root' | 'initializationOptions'> {
    id: string
}

interface ClientEntry extends SubscriptionLike {
    key: ClientKey
    client: Client
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

type ConfigurationUpdate = ConfigurationUpdateParams & MessageSource & PromiseCallback<void>

/** Options for creating the controller. */
export interface ControllerOptions<X extends Extension> extends Pick<ClientOptions, 'middleware'> {
    /** Creates transports to pass to the client to communicate with the given extension. */
    createMessageTransports: (
        extension: X,
        clientOptions: ClientOptions
    ) => MessageTransports | Promise<MessageTransports>

    /** Creates an error handler for a client for the given extension. */
    errorHandler?: (extension: X) => ErrorHandler

    /** Creates an initialization failure handler for a client for the given extension. */
    initializationFailedHandler?: (extension: X) => InitializationFailedHandler
}

/**
 * The controller for the environment.
 */
export class Controller<X extends Extension = Extension> implements Unsubscribable {
    private _environment = new BehaviorSubject<Environment<X>>(EMPTY_ENVIRONMENT)

    private _clients = new BehaviorSubject<ClientEntry[]>([])

    /** An observable that emits whenever the set of clients managed by this controller changes. */
    public get clients(): Observable<Client[]> {
        return this._clients.pipe(map(entries => entries.map(({ client }) => client)))
    }

    private subscriptions = new Subscription()

    /** The registries for various providers that expose extension functionality. */
    public readonly registries = new Registries()

    private readonly _logMessages = new Subject<LogMessageParams & MessageSource>()
    private readonly _showMessages = new Subject<ShowMessageParams & MessageSource>()
    private readonly _showMessageRequests = new Subject<ShowMessageRequest>()
    private readonly _configurationUpdates = new Subject<ConfigurationUpdate>()

    /** Log messages from extensions. */
    public readonly logMessages: Observable<LogMessageParams & MessageSource> = this._logMessages

    /** Messages from extensions intended for display to the user. */
    public readonly showMessages: Observable<ShowMessageParams & MessageSource> = this._showMessages

    /** Messages from extensions requesting a response from the user. */
    public readonly showMessageRequests: Observable<ShowMessageRequest> = this._showMessageRequests

    /** Configuration updates from extensions. */
    public readonly configurationUpdates: Observable<ConfigurationUpdate> = this._configurationUpdates

    constructor(private options: ControllerOptions<X>) {
        this.subscriptions.add(() => {
            for (const c of this._clients.value) {
                c.unsubscribe()
            }
        })
    }

    public setEnvironment(nextEnvironment: Environment<X>): void {
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
        for (const oldClient of this._clients.value) {
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
            unusedClient.unsubscribe()
        }
        // Create new clients.
        for (const { key } of newClients) {
            // Find the extension that this client is for.
            const extension = environment.extensions!.find(x => x.id === key.id)
            if (!extension) {
                throw new Error(`extension not found: ${key.id}`)
            }

            const clientOptions: ClientOptions = {
                ...this.options,
                root: key.root,
                initializationOptions: { ...key.initializationOptions }, // key is immutable so we can diff it
                documentSelector: ['*'],
                environment: this.environment,
                createMessageTransports: () => this.options.createMessageTransports(extension, clientOptions),
                initializationFailedHandler:
                    this.options.initializationFailedHandler && this.options.initializationFailedHandler(extension),
                errorHandler: this.options.errorHandler && this.options.errorHandler(extension),
            }
            const client = new Client(key.id, key.id, clientOptions)

            const settings = this._environment.pipe(
                map(({ extensions }) => (extensions ? extensions.find(x => x.id === key.id) : null)),
                filter((x): x is X => !!x),
                map(x => x.settings),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
            this.registerClientFeatures(client, settings)
            nextClients.push({
                key,
                client,
                ...client.start(), // SubscriptionLike
            })
        }
        this._clients.next(nextClients)
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
        client.registerFeature(new ExecuteCommandFeature(client, this.registries.commands))
        client.registerFeature(new TextDocumentDidOpenFeature(client))
        client.registerFeature(new TextDocumentDefinitionFeature(client, this.registries.textDocumentDefinition))
        client.registerFeature(
            new TextDocumentImplementationFeature(client, this.registries.textDocumentImplementation)
        )
        client.registerFeature(new TextDocumentReferencesFeature(client, this.registries.textDocumentReferences))
        client.registerFeature(
            new TextDocumentTypeDefinitionFeature(client, this.registries.textDocumentTypeDefinition)
        )
        client.registerFeature(new TextDocumentHoverFeature(client, this.registries.textDocumentHover))
        client.registerFeature(new TextDocumentStaticDecorationFeature(client, this.registries.textDocumentDecoration))
        client.registerFeature(new TextDocumentDynamicDecorationFeature(client, this.registries.textDocumentDecoration))
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
                    })
            )
        )
    }

    public readonly environment: ObservableEnvironment<X> = createObservableEnvironment<X>(this._environment)

    public set trace(value: Trace) {
        for (const client of this._clients.value) {
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
    if (!environment.root || !environment.extensions) {
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
