import { Observable, Subscription } from 'rxjs'
import { first, map } from 'rxjs/operators'
import uuidv4 from 'uuid/v4'
import { MessageType as RPCMessageType } from '../../jsonrpc2/messages'
import {
    ClientCapabilities,
    ConfigurationCascade,
    ConfigurationRequest,
    ConfigurationUpdateParams,
    ConfigurationUpdateRequest,
    DidChangeConfigurationNotification,
    InitializeParams,
    ServerCapabilities,
} from '../../protocol'
import { URI } from '../../types/textDocument'
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
        ensure(ensure(capabilities, 'workspace')!, 'didChangeConfiguration')!.dynamicRegistration = true
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
 * Support for the server requesting the client's configuration (workspace/configuration request to the client).
 *
 * @template C configuration cascade type
 */
export class ConfigurationFeature<C extends ConfigurationCascade> implements StaticFeature {
    constructor(private client: Client, private configurationCascade: Observable<C>) {}

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        capabilities.workspace = capabilities.workspace || {}
        capabilities.workspace!.configuration = true
    }

    public initialize(): void {
        this.client.onRequest(ConfigurationRequest.type, (params, token) => {
            const configuration: ConfigurationRequest.HandlerSignature = params =>
                this.configurationCascade
                    .pipe(
                        first(),
                        map(configurationCascade => {
                            const result: any[] = []
                            for (const item of params.items) {
                                result.push(
                                    this.getConfiguration(
                                        configurationCascade,
                                        item.scopeUri,
                                        item.section !== null ? item.section : undefined
                                    )
                                )
                            }
                            return result
                        })
                    )
                    .toPromise()
            return configuration(params, token)
        })
    }

    private getConfiguration(configurationCascade: C, resource: URI | undefined, section: string | undefined): C {
        if (resource) {
            throw new Error('configuration request: resource param is not supported')
        }
        if (section) {
            throw new Error('configuration request: section param is not supported')
        }
        // TODO(sqs): Support only returning partial settings (based on the resource/section args).
        // Also figure out in what cases it's OK for one extension to see another extension's
        // settings. In some cases this is dangerous because it would let extensions see the access
        // tokens, etc., configured for other extensions.
        return configurationCascade
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
