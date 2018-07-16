import { basename } from 'path'
import { BehaviorSubject, from, Observable, Subscription, SubscriptionLike, Unsubscribable } from 'rxjs'
import { filter, first, map, switchMap } from 'rxjs/operators'
import { ObservableEnvironment } from '../environment/environment'
import { MessageTransports } from '../jsonrpc2/connection'
import {
    GenericNotificationHandler,
    GenericRequestHandler,
    NotificationHandler,
    RequestHandler,
} from '../jsonrpc2/handlers'
import { Message, MessageType as RPCMessageType } from '../jsonrpc2/messages'
import { NotificationType, RequestType } from '../jsonrpc2/messages'
import { Trace, Tracer } from '../jsonrpc2/trace'
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
import { isFunction } from '../util'
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

    middleware?: Middleware

    /** Called to create the connection to the server. */
    createMessageTransports: () => MessageTransports | Promise<MessageTransports>

    environment: ObservableEnvironment
}

/** The client options, after defaults have been set that make certain fields required. */
interface ResolvedClientOptions extends ClientOptions {
    initializationFailedHandler: InitializationFailedHandler
    errorHandler: ErrorHandler
    middleware: Middleware
}

/** The possible states of a client. */
export enum ClientState {
    Initial,
    Starting,
    StartFailed,
    Running,
    Stopping,
    Stopped,
}

/**
 * The client communicates with a CXP server.
 */
export class Client implements Unsubscribable {
    public readonly clientOptions: ResolvedClientOptions

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

    private _trace: Trace = Trace.Off
    private _tracer: Tracer

    private subscriptions = new Subscription()

    public constructor(public readonly id: string, public readonly name: string, clientOptions: ClientOptions) {
        this.clientOptions = {
            ...clientOptions,
            initializationFailedHandler: clientOptions.initializationFailedHandler || (() => Promise.resolve(false)),
            errorHandler: clientOptions.errorHandler || new DefaultErrorHandler(),
            middleware: clientOptions.middleware || {},
        }

        this._tracer = {
            log: (message: string, data?: string) => {
                this.logTrace(message, data)
            },
        }
    }

    private isConnectionActive(): boolean {
        return this._state.value === ClientState.Running && this.connection !== null
    }

    public needsStop(): boolean {
        return this._state.value === ClientState.Starting || this._state.value === ClientState.Running
    }

    public start(): SubscriptionLike {
        this._state.next(ClientState.Starting)
        this.resolveConnection()
            .then(connection => {
                connection.listen()
                return this.initialize(connection)
            })
            .then(null, err => {
                this._state.next(ClientState.StartFailed)
                throw err
            })

        const c = this
        return {
            unsubscribe: () => {
                if (this.needsStop()) {
                    this.stop().then(null, err => console.error(err))
                }
            },
            get closed(): boolean {
                return c.needsStop()
            },
        }
    }

    private resolveConnection(): Promise<Connection> {
        if (!this.connectionPromise) {
            this.connectionPromise = this.createConnection()
        }
        return this.connectionPromise
    }

    private initialize(connection: Connection): Promise<void> {
        const initParams: InitializeParams = {
            root: this.clientOptions.root,
            capabilities: {},
            initializationOptions: isFunction(this.clientOptions.initializationOptions)
                ? this.clientOptions.initializationOptions()
                : this.clientOptions.initializationOptions,
            workspaceFolders: this.clientOptions.root
                ? [{ name: basename(this.clientOptions.root), uri: this.clientOptions.root }]
                : null,
            trace: Trace.toString(this._trace),
        }

        // Fill initialize params and client capabilities from features.
        for (const feature of this._features) {
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
                this.connection = connection
                this._initializeResult = result
                this._state.next(ClientState.Running)

                connection.onRequest(RegistrationRequest.type, params => this.handleRegistrationRequest(params))
                connection.onRequest(UnregistrationRequest.type, params => this.handleUnregistrationRequest(params))

                connection.sendNotification(InitializedNotification.type, {})

                // Initialize features.
                for (const feature of this._features) {
                    feature.initialize(result.capabilities, this.clientOptions.documentSelector)
                }
            })
            .then(null, err =>
                Promise.resolve(this.clientOptions.initializationFailedHandler(err)).then(restart => {
                    if (restart) {
                        return this.initialize(connection)
                    }
                    return this.stop()
                })
            )
    }

    private createConnection(): Promise<Connection> {
        return Promise.resolve(this.clientOptions.createMessageTransports()).then(transports =>
            createConnection(
                transports,
                (error, message, count) => this.handleConnectionError(error, message, count),
                () => this.handleConnectionClosed()
            )
        )
    }

    protected handleConnectionClosed(): void {
        // Check whether this is a normal shutdown in progress or the client stopped normally.
        if (this._state.value === ClientState.Stopping || this._state.value === ClientState.Stopped) {
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
        this.cleanUp()

        let action: Promise<CloseAction> = Promise.resolve(CloseAction.DoNotRestart)
        try {
            action = Promise.resolve(this.clientOptions.errorHandler.closed())
        } catch (error) {
            // Ignore sync errors from the error handler.
        }
        // tslint:disable-next-line:no-floating-promises
        action
            .catch(() => CloseAction.DoNotRestart) // ignore async errors from the error handler
            .then(action => {
                if (action === CloseAction.DoNotRestart) {
                    this._state.next(ClientState.Stopped)
                } else if (action === CloseAction.Restart) {
                    this._state.next(ClientState.Initial)
                    this.start()
                }
            })
    }

    private handleConnectionError(error: Error, message: Message | undefined, count: number | undefined): void {
        // Casts to any are required because the ErrorHandler interface is not using strictNullTypes.
        const action = this.clientOptions.errorHandler.error(error, message, count)
        if (action === ErrorAction.Shutdown) {
            this.stop().then(null, err => console.error(err))
        }
    }

    public stop(): Promise<void> {
        this._initializeResult = null
        if (!this.connectionPromise) {
            this._state.next(ClientState.Stopped)
            return Promise.resolve()
        }
        if (this._state.value === ClientState.Stopping && this.onStop) {
            return this.onStop
        }
        this._state.next(ClientState.Stopping)
        this.cleanUp()
        return (this.onStop = this.resolveConnection().then(connection =>
            connection.shutdown().then(() => {
                connection.exit()
                connection.unsubscribe()
                this._state.next(ClientState.Stopped)
                this.onStop = null
                this.connectionPromise = null
                this.connection = null
            })
        ))
    }

    private cleanUp(): void {
        for (const handler of this._dynamicFeatures.values()) {
            handler.unregisterAll()
        }
    }

    public sendRequest<P, R, E, RO>(type: RequestType<P, R, E, RO>, params: P): Promise<R>
    public sendRequest<R>(method: string, params?: any): Promise<R>
    public sendRequest<R>(type: string | RPCMessageType, ...params: any[]): Promise<R> {
        if (!this.isConnectionActive()) {
            throw new Error('connection is inactive')
        }
        return Promise.resolve(this.connection!.sendRequest<R>(type, ...params))
    }

    public onRequest<P, R, E, RO>(type: RequestType<P, R, E, RO>, handler: RequestHandler<P, R, E>): void
    public onRequest<R, E>(method: string, handler: GenericRequestHandler<R, E>): void
    public onRequest<R, E>(type: string | RPCMessageType, handler: GenericRequestHandler<R, E>): void {
        if (!this.isConnectionActive()) {
            throw new Error('connection is inactive')
        }
        this.connection!.onRequest(type, handler)
    }

    public sendNotification<P, RO>(type: NotificationType<P, RO>, params?: P): void
    public sendNotification(method: string, params?: any): void
    public sendNotification<P>(type: string | RPCMessageType, params?: P): void {
        if (!this.isConnectionActive()) {
            throw new Error('connection is inactive')
        }
        this.connection!.sendNotification(type, params)
    }

    public onNotification<P, RO>(type: NotificationType<P, RO>, handler: NotificationHandler<P>): void
    public onNotification(method: string, handler: GenericNotificationHandler): void
    public onNotification(type: string | RPCMessageType, handler: GenericNotificationHandler): void {
        if (!this.isConnectionActive()) {
            throw new Error('connection is inactive')
        }
        this.connection!.onNotification(type, handler)
    }

    public set trace(value: Trace) {
        this._trace = value
        this._state
            .pipe(
                filter(state => state === ClientState.Running),
                first(),
                switchMap(() => from(this.resolveConnection())),
                map(connection => connection.trace(value, this._tracer))
            )
            .subscribe()
    }

    private logTrace(message: string, data?: any): void {
        console.info(message, data)
    }

    private _features: (StaticFeature | DynamicFeature<any>)[] = []
    private readonly _method2Message: Map<string, RPCMessageType> = new Map<string, RPCMessageType>()
    private readonly _dynamicFeatures: Map<string, DynamicFeature<any>> = new Map<string, DynamicFeature<any>>()

    public registerFeature(feature: StaticFeature | DynamicFeature<any>): void {
        this._features.push(feature)
        if (DynamicFeature.is(feature)) {
            const messages = feature.messages
            if (Array.isArray(messages)) {
                for (const message of messages) {
                    this._method2Message.set(message.method, message)
                    this._dynamicFeatures.set(message.method, feature)
                }
            } else {
                this._method2Message.set(messages.method, messages)
                this._dynamicFeatures.set(messages.method, feature)
            }
        }
    }

    private handleRegistrationRequest(params: RegistrationParams): void {
        for (const registration of params.registrations) {
            const feature = this._dynamicFeatures.get(registration.method)
            if (!feature) {
                throw new Error(`No feature implementation for ${registration.method} found. Registration failed.`)
            }
            const options = registration.registerOptions || {}
            options.documentSelector = options.documentSelector || this.clientOptions.documentSelector
            const data: RegistrationData<any> = {
                id: registration.id,
                registerOptions: options,
            }
            feature.register(this._method2Message.get(registration.method)!, data)
        }
    }

    private handleUnregistrationRequest(params: UnregistrationParams): void {
        for (const unregistration of params.unregisterations) {
            const feature = this._dynamicFeatures.get(unregistration.method)
            if (!feature) {
                throw new Error(`No feature implementation for ${unregistration.method} found. Unregistration failed.`)
            }
            feature.unregister(unregistration.id)
        }
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
