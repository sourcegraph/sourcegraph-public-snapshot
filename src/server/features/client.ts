import { Unsubscribable } from 'rxjs'
import uuidv4 from 'uuid/v4'
import { MessageType as RPCMessageType, NotificationType, RequestType } from '../../jsonrpc2/messages'
import {
    InitializeParams,
    Registration,
    RegistrationParams,
    RegistrationRequest,
    ServerCapabilities,
    Unregistration,
    UnregistrationParams,
} from '../../protocol'
import { UnregistrationRequest } from '../../protocol/registration'
import { IConnection } from '../server'
import { Remote } from './common'

/**
 * A bulk registration manages n single registration to be able to register
 * for n notifications or requests using one register request.
 */
export interface BulkRegistration {
    /**
     * Adds a single registration.
     * @param type the notification type to register for.
     * @param registerParams special registration parameters.
     */
    add<P, RO>(type: NotificationType<P, RO>, registerParams: RO): void

    /**
     * Adds a single registration.
     * @param type the request type to register for.
     * @param registerParams special registration parameters.
     */
    add<P, R, E, RO>(type: RequestType<P, R, E, RO>, registerParams: RO): void
}

class BulkRegistrationImpl implements BulkRegistration {
    private _registrations: Registration[] = []
    private _registered: Set<string> = new Set<string>()

    public add<RO>(type: string | RPCMessageType, registerOptions?: RO): void {
        const method = typeof type === 'string' ? type : type.method
        if (this._registered.has(method)) {
            throw new Error(`${method} is already added to this registration`)
        }
        const id = uuidv4()
        this._registrations.push({
            id,
            method,
            registerOptions: registerOptions || {},
        })
        this._registered.add(method)
    }

    public asRegistrationParams(): RegistrationParams {
        return {
            registrations: this._registrations,
        }
    }
}

export namespace BulkRegistration {
    /**
     * Creates a new bulk registration.
     * @return an empty bulk registration.
     */
    export function create(): BulkRegistration {
        return new BulkRegistrationImpl()
    }
}

/**
 * A `BulkUnregistration` manages n unregistrations.
 */
export interface BulkUnregistration extends Unsubscribable {
    /**
     * Unsubscribes a single registration. It will be removed from the
     * `BulkUnregistration`.
     */
    unsubscribeSingle(arg: string | RPCMessageType): boolean
}

class BulkUnregistrationImpl implements BulkUnregistration {
    private _unregistrations: Map<string, Unregistration> = new Map<string, Unregistration>()

    constructor(private _connection: IConnection | undefined, unregistrations: Unregistration[]) {
        for (const unregistration of unregistrations) {
            this._unregistrations.set(unregistration.method, unregistration)
        }
    }

    public get isAttached(): boolean {
        return !!this._connection
    }

    public attach(connection: IConnection): void {
        this._connection = connection
    }

    public add(unregistration: Unregistration): void {
        this._unregistrations.set(unregistration.method, unregistration)
    }

    public unsubscribe(): void {
        const unregistrations: Unregistration[] = []
        for (const unregistration of this._unregistrations.values()) {
            unregistrations.push(unregistration)
        }
        const params: UnregistrationParams = {
            unregisterations: unregistrations,
        }
        this._connection!.sendRequest(UnregistrationRequest.type, params).then(undefined, _error => {
            this._connection!.console.info(`Bulk unregistration failed.`)
        })
    }

    public unsubscribeSingle(arg: string | RPCMessageType): boolean {
        const method = typeof arg === 'string' ? arg : arg.method

        const unregistration = this._unregistrations.get(method)
        if (!unregistration) {
            return false
        }

        const params: UnregistrationParams = {
            unregisterations: [unregistration],
        }
        this._connection!.sendRequest(UnregistrationRequest.type, params).then(
            () => {
                this._unregistrations.delete(method)
            },
            _error => {
                this._connection!.console.info(`Unregistering request handler for ${unregistration.id} failed.`)
            }
        )
        return true
    }
}

export namespace BulkUnregistration {
    export function create(): BulkUnregistration {
        return new BulkUnregistrationImpl(undefined, [])
    }
}

/**
 * Interface to register and unregister `listeners` on the client / tools side.
 */
export interface RemoteClient extends Remote {
    /**
     * Registers a listener for the given notification.
     * @param type the notification type to register for.
     * @param registerParams special registration parameters.
     * @return a `Unsubscribable` to unregister the listener again.
     */
    register<P, RO>(type: NotificationType<P, RO>, registerParams?: RO): Promise<Unsubscribable>

    /**
     * Registers a listener for the given notification.
     * @param unregisteration the unregistration to add a corresponding unregister action to.
     * @param type the notification type to register for.
     * @param registerParams special registration parameters.
     * @return the updated unregistration.
     */
    register<P, RO>(
        unregisteration: BulkUnregistration,
        type: NotificationType<P, RO>,
        registerParams?: RO
    ): Promise<BulkUnregistration>

    /**
     * Registers a listener for the given request.
     * @param type the request type to register for.
     * @param registerParams special registration parameters.
     * @return a `Unsubscribable` to unregister the listener again.
     */
    register<P, R, E, RO>(type: RequestType<P, R, E, RO>, registerParams?: RO): Promise<Unsubscribable>

    /**
     * Registers a listener for the given request.
     * @param unregisteration the unregistration to add a corresponding unregister action to.
     * @param type the request type to register for.
     * @param registerParams special registration parameters.
     * @return the updated unregistration.
     */
    register<P, R, E, RO>(
        unregisteration: BulkUnregistration,
        type: RequestType<P, R, E, RO>,
        registerParams?: RO
    ): Promise<BulkUnregistration>

    /**
     * Registers a set of listeners.
     * @param registrations the bulk registration
     * @return a `Unsubscribable` to unregister the listeners again.
     */
    register(registrations: BulkRegistration): Promise<BulkUnregistration>
}

export class RemoteClientImpl implements RemoteClient {
    private _connection?: IConnection

    public attach(connection: IConnection): void {
        this._connection = connection
    }

    public get connection(): IConnection {
        if (!this._connection) {
            throw new Error('Remote is not attached to a connection yet.')
        }
        return this._connection
    }

    public initialize(_params: InitializeParams): void {
        /* noop */
    }

    public fillServerCapabilities(_capabilities: ServerCapabilities): void {
        /* noop */
    }

    public register(
        typeOrRegistrations: string | RPCMessageType | BulkRegistration | BulkUnregistration,
        registerOptionsOrType?: string | RPCMessageType | any,
        registerOptions?: any
    ): Promise<any> /* Promise<Unsubscribable | BulkUnregistration> */ {
        if (typeOrRegistrations instanceof BulkRegistrationImpl) {
            return this.registerMany(typeOrRegistrations)
        } else if (typeOrRegistrations instanceof BulkUnregistrationImpl) {
            return this.registerSingle1(
                typeOrRegistrations as BulkUnregistrationImpl,
                registerOptionsOrType as string | RPCMessageType,
                registerOptions
            )
        } else {
            return this.registerSingle2(typeOrRegistrations as string | RPCMessageType, registerOptionsOrType)
        }
    }

    private registerSingle1(
        unregistration: BulkUnregistrationImpl,
        type: string | RPCMessageType,
        registerOptions: any
    ): Promise<Unsubscribable> {
        const method = typeof type === 'string' ? type : type.method
        const id = uuidv4()
        const params: RegistrationParams = {
            registrations: [{ id, method, registerOptions: registerOptions || {} }],
        }
        if (!unregistration.isAttached) {
            unregistration.attach(this.connection)
        }
        return this.connection.sendRequest(RegistrationRequest.type, params).then(
            _result => {
                unregistration.add({ id, method })
                return unregistration
            },
            _error => {
                this.connection.console.info(`Registering request handler for ${method} failed.`)
                return Promise.reject(_error)
            }
        )
    }

    private registerSingle2(type: string | RPCMessageType, registerOptions: any): Promise<Unsubscribable> {
        const method = typeof type === 'string' ? type : type.method
        const id = uuidv4()
        const params: RegistrationParams = {
            registrations: [{ id, method, registerOptions: registerOptions || {} }],
        }
        return this.connection.sendRequest(RegistrationRequest.type, params).then(
            _result => ({ unsubscribe: () => this.unregisterSingle(id, method) }),
            _error => {
                this.connection.console.info(`Registering request handler for ${method} failed.`)
                return Promise.reject(_error)
            }
        )
    }

    private unregisterSingle(id: string, method: string): Promise<void> {
        const params: UnregistrationParams = {
            unregisterations: [{ id, method }],
        }

        return this.connection.sendRequest(UnregistrationRequest.type, params).then(undefined, _error => {
            this.connection.console.info(`Unregistering request handler for ${id} failed.`)
        })
    }

    private registerMany(registrations: BulkRegistrationImpl): Promise<BulkUnregistration> {
        const params = registrations.asRegistrationParams()
        return this.connection.sendRequest(RegistrationRequest.type, params).then(
            () =>
                new BulkUnregistrationImpl(
                    this._connection,
                    params.registrations.map(registration => ({ id: registration.id, method: registration.method }))
                ),
            _error => {
                this.connection.console.info(`Bulk registration failed.`)
                return Promise.reject(_error)
            }
        )
    }
}
