import { Observable, Subscription } from 'rxjs'
import { first } from 'rxjs/operators'
import uuidv4 from 'uuid/v4'
import { MessageType as RPCMessageType } from '../../jsonrpc2/messages'
import {
    ClientCapabilities,
    ConfigurationCascade,
    ConfigurationUpdateParams,
    ConfigurationUpdateRequest,
    DidChangeConfigurationNotification,
    InitializeParams,
    ServerCapabilities,
} from '../../protocol'
import { Client } from '../client'
import { DynamicFeature, ensure, RegistrationData, StaticFeature } from './common'

/**
 * Support for configuration settings managed by the client (workspace/didChangeConfiguration notifications to the
 * server).
 *
 * @template C configuration cascade type
 */
export class ConfigurationChangeNotificationFeature<C extends ConfigurationCascade>
    implements DynamicFeature<undefined> {
    private subscriptionsByID = new Map<string, Subscription>()

    constructor(private client: Client, private configurationCascade: Observable<C>) {}

    public readonly messages = DidChangeConfigurationNotification.type

    public fillInitializeParams(params: InitializeParams): void {
        // This runs synchronously because this.configurationCascade's root source is a BehaviorSubject (which has
        // an initial value). Confirm it is synchronous just in case, because a bug here would be hard to diagnose.
        let sync = false
        this.configurationCascade
            .pipe(first())
            .subscribe(configurationCascade => {
                params.configurationCascade = configurationCascade
                sync = true
            })
            .unsubscribe()
        if (!sync) {
            throw new Error('configuration is not immediately available')
        }
    }

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        ensure(ensure(capabilities, 'configuration')!, 'didChangeConfiguration')!.dynamicRegistration = true
    }

    public initialize(capabilities: ServerCapabilities): void {
        this.register(this.messages, { id: uuidv4(), registerOptions: undefined })
    }

    public register(message: RPCMessageType, data: RegistrationData<undefined>): void {
        if (this.subscriptionsByID.has(data.id)) {
            throw new Error(`registration already exists with ID ${data.id}`)
        }
        this.subscriptionsByID.set(
            data.id,
            this.configurationCascade.subscribe(configurationCascade =>
                this.client.sendNotification(DidChangeConfigurationNotification.type, { configurationCascade })
            )
        )
    }

    public unregister(id: string): void {
        const sub = this.subscriptionsByID.get(id)
        if (!sub) {
            throw new Error(`no registration with ID ${id}`)
        }
        this.subscriptionsByID.delete(id)
    }

    public unregisterAll(): void {
        for (const sub of this.subscriptionsByID.values()) {
            sub.unsubscribe()
        }
        this.subscriptionsByID.clear()
    }
}

/**
 * Support for the server updating the client's configuration (configuration/update request to the
 * client).
 */
export class ConfigurationUpdateFeature implements StaticFeature {
    constructor(
        private client: Client,
        /** Called when the client receives a configuration/update request. */
        private update: (params: ConfigurationUpdateParams) => Promise<void>
    ) {}

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        ensure(capabilities, 'configuration')!.update = true
    }

    public initialize(): void {
        this.client.onRequest(ConfigurationUpdateRequest.type, this.update)
    }
}
