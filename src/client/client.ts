import { basename } from 'path'
import { BehaviorSubject, Observable, Unsubscribable } from 'rxjs'
import { MessageTransports } from '../jsonrpc2/connection'
import {
    GenericNotificationHandler,
    GenericRequestHandler,
    NotificationHandler,
    RequestHandler,
} from '../jsonrpc2/handlers'
import { Message, MessageType as RPCMessageType } from '../jsonrpc2/messages'
import { NotificationType, RequestType } from '../jsonrpc2/messages'
import { noopTracer, Trace, Tracer } from '../jsonrpc2/trace'
import {
    InitializedNotification,
    InitializeParams,
    InitializeResult,
    RegistrationParams,
    RegistrationRequest,
    UnregistrationParams,
    UnregistrationRequest,
} from '../protocol'
import { DocumentSelector } from '../types/document'
import { URI } from '../types/textDocument'
import { isFunction, tryCatchPromise } from '../util'
import { Connection, createConnection } from './connection'
import {
    CloseAction,
    DefaultErrorHandler,
    ErrorAction,
    ErrorHandler,
    InitializationFailedHandler,
} from './errorHandler'
import { DynamicFeature, RegistrationData, StaticFeature } from './features/common'
import { Middleware } from './middleware'

/** Options for creating a new client. */
export interface ClientOptions {
    root: URI | null
    documentSelector?: DocumentSelector
    initializationOptions?: any | (() => any)

    /** Called when initialization fails to determine how to proceed. */
    initializationFailedHandler?: InitializationFailedHandler

    /** Called when an error or close occurs to determine how to proceed. */
    errorHandler?: ErrorHandler

    middleware?: Readonly<Middleware>

    /** Called to create the connection to the server. */
    createMessageTransports: () => MessageTransports | Promise<MessageTransports>

    /** Trace log level. The tracer must also be set. */
    trace?: Trace

    /** Logs messages sent to and received from the server, and trace log messages sent from the server. */
    tracer?: Tracer
}

/** The client options, after defaults have been set that make certain fields required. */
interface ResolvedClientOptions extends Pick<ClientOptions, Exclude<keyof ClientOptions, 'trace'>> {
    initializationFailedHandler: InitializationFailedHandler
    errorHandler: ErrorHandler
    middleware: Readonly<Middleware>
    tracer: Tracer
}

/** The possible states of a client. */
export enum ClientState {
    /** The initial state of the client. It has never been activated. */
    Initial,

    /** The client is establishing the connection to the server and sending the "initialize" message. */
    Connecting,

    /**
     * The connection is established and the client is waiting for and handling the server's "initialize" result.
     */
    Initializing,

    /** The client encountered an error while activating. */
    ActivateFailed,

    /** The client has finished initialization and is in operation. */
    Active,

    /** The client is gracefully shutting down the connection. */
    ShuttingDown,

    /** The client is deactivated (after having previously been activated). */
    Stopped,
}

/**
 * The client communicates with a CXP server.
 */
export class Client implements Unsubscribable {
    public readonly options: Readonly<ResolvedClientOptions>

    private _initializeResult: InitializeResult | null = null
    public get initializeResult(): InitializeResult | null {
        return this._initializeResult
    }

    private _state = new BehaviorSubject<ClientState>(ClientState.Initial)
    public get state(): Observable<ClientState> {
        return this._state
    }

    private connectionPromise: Promise<Connection> | null = null
    private connection: Connection | null = null

    private onStop: Promise<void> | null = null

    private _trace: Trace

    public constructor(public readonly id: string, public readonly name: string, { trace, ...options }: ClientOptions) {
        this.options = {
            ...options,
            initializationFailedHandler: options.initializationFailedHandler || (() => Promise.resolve(false)),
            errorHandler: options.errorHandler || new DefaultErrorHandler(),
            middleware: options.middleware || {},
            tracer: options.tracer || noopTracer,
        }
        this._trace = trace || Trace.Off
    }

    private get isConnectionActive(): boolean {
        return (
            (this._state.value === ClientState.Initializing || this._state.value === ClientState.Active) &&
            this.connection !== null
        )
    }

    public needsStop(): boolean {
        return (
            this._state.value === ClientState.Connecting ||
            this._state.value === ClientState.Initializing ||
            this._state.value === ClientState.Active
        )
    }

    /**
     * Activates the client, which causes it to start connecting (and to reestablish the connection when it drops,
     * as directed by the initializationFailedHandler).
     *
     * To watch client state, use Client#state. To log client errors, provide an initializationFailedHandler and
     * errorHandler in ClientOptions.
     */
    public activate(): void {
        // Callers should subscribe to Client#state instead of awaiting the activation.
        this.activateAndWait().then(null, () => void 0)
    }

    /**
     * Activates the client and returns a promise that resolves when the initial activation finishes (or fails and
     * no retry will occur). Used by tests.
     */
    protected activateAndWait(): Promise<void> {
        this._state.next(ClientState.Connecting)
        let activateConnection: Connection | null = null // track so we know if we're dealing with the same value upon error
        return this.resolveConnection()
            .then(connection => {
                activateConnection = connection
                connection.listen()
                return this.initialize(connection)
            })
            .then(null, err => {
                console.error(err)

                // Only update state if it pertains to the same connection we started with.
                if (activateConnection === this.connection && this._state.value !== ClientState.Stopped) {
                    this._state.next(ClientState.ActivateFailed)
                }
            })
    }

    private resolveConnection(): Promise<Connection> {
        if (!this.connectionPromise) {
            this.connectionPromise = tryCatchPromise(this.options.createMessageTransports).then(transports =>
                createConnection(
                    transports,
                    (error, message, count) => this.handleConnectionError(error, message, count),
                    () => this.handleConnectionClosed()
                )
            )
        }
        return this.connectionPromise
    }

    private initialize(connection: Connection): Promise<void> {
        connection.trace(this._trace, this.options.tracer)

        const initParams: InitializeParams = {
            root: this.options.root,
            capabilities: {},
            configurationCascade: { merged: {} },
            initializationOptions: isFunction(this.options.initializationOptions)
                ? this.options.initializationOptions()
                : this.options.initializationOptions,
            workspaceFolders: this.options.root
                ? [{ name: basename(this.options.root), uri: this.options.root }]
                : null,
            trace: Trace.toString(this._trace),
        }

        // Fill initialize params and client capabilities from features.
        for (const feature of this.features) {
            if (isFunction(feature.fillClientCapabilities)) {
                feature.fillClientCapabilities(initParams.capabilities)
            }
            if (isFunction(feature.fillInitializeParams)) {
                feature.fillInitializeParams(initParams)
            }
        }

        return connection
            .initialize(initParams)
            .then(result => {
                // If this client was stopped during the initialize call, don't continue'
                if (this._state.value === ClientState.ShuttingDown || this._state.value === ClientState.Stopped) {
                    return
                }

                this.connection = connection
                this._initializeResult = result

                this._state.next(ClientState.Initializing)

                connection.onRequest(RegistrationRequest.type, params => this.handleRegistrationRequest(params))
                connection.onRequest(UnregistrationRequest.type, params => this.handleUnregistrationRequest(params))

                connection.sendNotification(InitializedNotification.type, {})

                // Initialize features.
                for (const feature of this.features) {
                    feature.initialize(result.capabilities, this.options.documentSelector)
                }

                this._state.next(ClientState.Active)
            })
            .then(null, err =>
                Promise.resolve(this.options.initializationFailedHandler(err)).then(reinitialize => {
                    if (reinitialize) {
                        return this.initialize(connection)
                    }
                    if (connection) {
                        connection.unsubscribe()
                    }
                    return this.stopAtState(ClientState.ActivateFailed)
                })
            )
    }

    protected handleConnectionClosed(): void {
        // Check whether this is a normal shutdown in progress or the client stopped normally.
        if (
            this._state.value === ClientState.ShuttingDown ||
            this._state.value === ClientState.Stopped ||
            this._state.value === ClientState.ActivateFailed
        ) {
            return
        }
        try {
            if (this.connection) {
                this.connection.unsubscribe()
            }
        } catch (error) {
            // Unsubscribing a connection could fail if error cases.
        }

        this.connectionPromise = null
        this.connection = null
        this.unregisterFeatures()

        let action: Promise<CloseAction> = Promise.resolve(CloseAction.DoNotReconnect)
        try {
            action = Promise.resolve(this.options.errorHandler.closed())
        } catch (error) {
            // Ignore sync errors from the error handler.
        }
        // tslint:disable-next-line:no-floating-promises
        action
            .catch(() => CloseAction.DoNotReconnect) // ignore async errors from the error handler
            .then(action => {
                if (action === CloseAction.DoNotReconnect) {
                    this._state.next(ClientState.Stopped)
                } else if (action === CloseAction.Reconnect) {
                    this._state.next(ClientState.Initial)
                    this.activate()
                }
            })
    }

    private handleConnectionError(error: Error, message: Message | undefined, count: number | undefined): void {
        const action = this.options.errorHandler.error(error, message, count)
        if (action === ErrorAction.ShutDown) {
            this.stopAtState(this.isConnectionActive ? ClientState.Stopped : ClientState.ActivateFailed).then(
                null,
                err => console.error(err)
            )
        }
    }

    /**
     * Stops the client, which causes it to gracefully shut down the current connection (if any) and remain
     * disconnected until a subsequent call to Client#activate.
     *
     * @returns a promise that resolves when shutdown completes, or immediately if the client is not connected
     */
    public stop(): Promise<void> {
        return this.stopAtState(ClientState.Stopped)
    }

    private stopAtState(endState: ClientState.Stopped | ClientState.ActivateFailed): Promise<void> {
        this._initializeResult = null
        if (!this.connectionPromise) {
            this._state.next(ClientState.Stopped)
            return Promise.resolve()
        }
        if (this._state.value === ClientState.ShuttingDown && this.onStop) {
            return this.onStop
        }

        const closeConnection = (connection: Connection) => {
            // It's possible for the connection to be alive and this.isConnectionActive === false (e.g., if we're
            // still waiting to hear back from the server), so make sure we close it.
            if (connection) {
                connection.unsubscribe()
            }

            if (connection !== this.connection) {
                // Another connection was created while we were preparing this one to close. Don't modify any state
                // because the state reflects the new connection now.
                return
            }

            if (this._state.value !== endState) {
                this._state.next(endState)
            }
            this.onStop = null
            this.connectionPromise = null
            this.connection = null
        }

        // If we are connected to a server, then shut down gracefully. Otherwise (if we aren't connected to a
        // server, including if the connection never succeeded), just close the connection (if any) immediately.
        const wasConnectionActive = this.isConnectionActive
        if (this.isConnectionActive) {
            this._state.next(ClientState.ShuttingDown)
        } else {
            this._state.next(endState)
        }
        this.unregisterFeatures()
        if (wasConnectionActive) {
            // Shut down gracefully before closing the connection.
            const connection = this.connection!
            return (this.onStop = connection
                .shutdown()
                .catch(() => {
                    /* Ignore shutdown errors from server. */
                })
                .then(() => {
                    connection.exit()
                    closeConnection(connection)
                }))
        }
        // Otherwise, just close the connection.
        closeConnection(this.connection!)
        return Promise.resolve()
    }

    private unregisterFeatures(): void {
        for (const feature of this.features) {
            if (DynamicFeature.is(feature)) {
                feature.unregisterAll()
            } else if (feature.deinitialize) {
                feature.deinitialize()
            }
        }
    }

    public sendRequest<P, R, E, RO>(type: RequestType<P, R, E, RO>, params: P): Promise<R>
    public sendRequest<R>(method: string, params?: any): Promise<R>
    public sendRequest<R>(type: string | RPCMessageType, ...params: any[]): Promise<R> {
        if (!this.isConnectionActive) {
            throw new Error('connection is inactive')
        }
        return Promise.resolve(this.connection!.sendRequest<R>(type, ...params))
    }

    public onRequest<P, R, E, RO>(type: RequestType<P, R, E, RO>, handler: RequestHandler<P, R, E>): void
    public onRequest<R, E>(method: string, handler: GenericRequestHandler<R, E>): void
    public onRequest<R, E>(type: string | RPCMessageType, handler: GenericRequestHandler<R, E>): void {
        if (!this.isConnectionActive) {
            throw new Error('connection is inactive')
        }
        this.connection!.onRequest(type, handler)
    }

    public sendNotification<P, RO>(type: NotificationType<P, RO>, params?: P): void
    public sendNotification(method: string, params?: any): void
    public sendNotification<P>(type: string | RPCMessageType, params?: P): void {
        if (!this.isConnectionActive) {
            throw new Error('connection is inactive')
        }
        this.connection!.sendNotification(type, params)
    }

    public onNotification<P, RO>(type: NotificationType<P, RO>, handler: NotificationHandler<P>): void
    public onNotification(method: string, handler: GenericNotificationHandler): void
    public onNotification(type: string | RPCMessageType, handler: GenericNotificationHandler): void {
        if (!this.isConnectionActive) {
            throw new Error('connection is inactive')
        }
        this.connection!.onNotification(type, handler)
    }

    public get trace(): Trace {
        return this._trace
    }

    public set trace(value: Trace) {
        this._trace = value
        if (this.connection) {
            this.connection.trace(value, this.options.tracer)
        }
    }

    protected readonly features: (StaticFeature | DynamicFeature<any>)[] = []
    private readonly _method2Message: Map<string, RPCMessageType> = new Map<string, RPCMessageType>()
    private readonly _dynamicFeatures: Map<string, DynamicFeature<any>> = new Map<string, DynamicFeature<any>>()

    public registerFeature(feature: StaticFeature | DynamicFeature<any>): void {
        if (DynamicFeature.is(feature)) {
            const messages = Array.isArray(feature.messages) ? feature.messages : [feature.messages]
            for (const message of messages) {
                if (this._method2Message.has(message.method)) {
                    throw new Error(
                        `dynamic feature is already registered for method ${JSON.stringify(message.method)}`
                    )
                }
                this._method2Message.set(message.method, message)
                this._dynamicFeatures.set(message.method, feature)
            }
        }
        this.features.push(feature)
    }

    protected handleRegistrationRequest(params: RegistrationParams): void {
        for (const registration of params.registrations) {
            const feature = this._dynamicFeatures.get(registration.method)
            if (!feature) {
                throw new Error(`dynamic feature not found: ${JSON.stringify(registration.method)}`)
            }
            const options = registration.registerOptions || {}
            if (!options.documentSelector && this.options.documentSelector) {
                options.documentSelector = this.options.documentSelector
            }
            const data: RegistrationData<any> = {
                id: registration.id,
                registerOptions: options,
                overwriteExisting: registration.overwriteExisting,
            }
            feature.register(this._method2Message.get(registration.method)!, data)
        }
    }

    protected handleUnregistrationRequest(params: UnregistrationParams): void {
        for (const unregistration of params.unregisterations) {
            const feature = this._dynamicFeatures.get(unregistration.method)
            if (!feature) {
                throw new Error(`dynamic feature not found: ${JSON.stringify(unregistration.method)}`)
            }
            feature.unregister(unregistration.id)
        }
    }

    public unsubscribe(): void {
        // Immediately unregister feature registrations, even before the connection shuts down, since calling
        // unsubscribe is evidence that the consumer doesn't want this client's data anymore.
        this.unregisterFeatures()

        if (this.needsStop()) {
            this.stop().then(null, err => console.error(err))
        }
    }
}
