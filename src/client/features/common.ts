import { Subscription, TeardownLogic, Unsubscribable } from 'rxjs'
import { MessageType as RPCMessageType } from '../../jsonrpc2/messages'
import {
    ClientCapabilities,
    InitializeParams,
    ServerCapabilities,
    TextDocumentRegistrationOptions,
} from '../../protocol'
import { DocumentSelector } from '../../types/document'
import { isFunction } from '../../util'
import { Client } from '../client'

export interface RegistrationData<T> {
    id: string
    registerOptions: T
}

export interface StaticFeature {
    fillInitializeParams?: (params: InitializeParams) => void
    fillClientCapabilities?: (capabilities: ClientCapabilities) => void
    initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector | undefined): void
}

export interface DynamicFeature<T> extends Unsubscribable {
    messages: RPCMessageType | RPCMessageType[]
    fillInitializeParams?: (params: InitializeParams) => void
    fillClientCapabilities(capabilities: ClientCapabilities): void
    initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector | undefined): void
    register(message: RPCMessageType, data: RegistrationData<T>): void
    unregister(id: string): void
}

export namespace DynamicFeature {
    export function is<T>(value: any): value is DynamicFeature<T> {
        const candidate: DynamicFeature<T> = value
        return (
            candidate &&
            // tslint:disable-next-line:no-unbound-method
            isFunction(candidate.register) &&
            // tslint:disable-next-line:no-unbound-method
            isFunction(candidate.unregister) &&
            // tslint:disable-next-line:no-unbound-method
            isFunction(candidate.unsubscribe) &&
            candidate.messages !== void 0
        )
    }
}

export abstract class TextDocumentFeature<T extends TextDocumentRegistrationOptions> implements DynamicFeature<T> {
    private subscriptions = new Subscription()
    private subscriptionsByID = new Map<string, Subscription>()

    constructor(protected client: Client, private _message: RPCMessageType) {}

    public get messages(): RPCMessageType {
        return this._message
    }

    public register(message: RPCMessageType, data: RegistrationData<T>): void {
        if (message.method !== this.messages.method) {
            throw new Error(
                `Register called on wrong feature. Requested ${message.method} but reached feature ${
                    this.messages.method
                }`
            )
        }
        if (!data.registerOptions.documentSelector) {
            return
        }
        const provider = this.registerProvider(data.registerOptions)
        if (provider) {
            const sub = this.subscriptions.add(provider)
            this.subscriptionsByID.set(data.id, sub)
        }
    }

    protected abstract registerProvider(options: T): TeardownLogic

    public abstract fillClientCapabilities(capabilities: ClientCapabilities): void

    public abstract initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector): void

    public unregister(id: string): void {
        const sub = this.subscriptionsByID.get(id)
        if (sub) {
            this.subscriptions.remove(sub)
        }
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
        this.subscriptionsByID.clear()
    }
}

export function ensure<T, K extends keyof T>(target: T, key: K): T[K] {
    if (target[key] === void 0) {
        target[key] = {} as any
    }
    return target[key]
}
