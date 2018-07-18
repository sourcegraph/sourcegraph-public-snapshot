import { Observable, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'
import uuidv4 from 'uuid/v4'
import { TextDocument } from 'vscode-languageserver-types'
import { ObservableEnvironment } from '../../environment/environment'
import { MessageType as RPCMessageType, NotificationType } from '../../jsonrpc2/messages'
import {
    ClientCapabilities,
    DidOpenTextDocumentNotification,
    DidOpenTextDocumentParams,
    ServerCapabilities,
    TextDocumentRegistrationOptions,
} from '../../protocol'
import { DocumentSelector } from '../../types/document'
import { NextSignature } from '../../types/middleware'
import { match } from '../../types/textDocument'
import { Client } from '../client'
import { DynamicFeature, ensure, RegistrationData } from './common'

type CreateParamsSignature<E, P> = (data: E) => P

export abstract class TextDocumentNotificationFeature<P, E> implements DynamicFeature<TextDocumentRegistrationOptions> {
    private subscription: Subscription | null = null

    protected selectors = new Map<string, DocumentSelector>()

    constructor(
        protected client: Client,
        protected observable: Observable<E>,
        protected type: NotificationType<P, TextDocumentRegistrationOptions>,
        protected middleware: NextSignature<E, void> | undefined,
        protected createParams: CreateParamsSignature<E, P>,
        protected selectorFilter?: (selectors: IterableIterator<DocumentSelector>, data: E) => boolean
    ) {}

    public abstract messages: RPCMessageType | RPCMessageType[]

    public abstract fillClientCapabilities(capabilities: ClientCapabilities): void

    public abstract initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector | undefined): void

    public register(_message: RPCMessageType, data: RegistrationData<TextDocumentRegistrationOptions>): void {
        if (!data.registerOptions.documentSelector) {
            return
        }
        if (this.selectors.has(data.id)) {
            throw new Error(`registration already exists with ID ${data.id}`)
        }
        this.selectors.set(data.id, data.registerOptions.documentSelector)
        if (!this.subscription) {
            this.subscription = this.observable.subscribe(data => this.callback(data))
        }
    }

    private callback(data: E): void {
        if (!this.selectorFilter || this.selectorFilter(this.selectors.values(), data)) {
            if (this.middleware) {
                this.middleware(data, data => this.client.sendNotification(this.type, this.createParams(data)))
            } else {
                this.client.sendNotification(this.type, this.createParams(data))
            }
            this.notificationSent(data)
        }
    }

    protected notificationSent(_data: E): void {
        /* noop */
    }

    public unregister(id: string): void {
        if (!this.selectors.delete(id)) {
            throw new Error(`no registration with ID ${id}`)
        }
        this.selectors.delete(id)
        if (this.selectors.size === 0 && this.subscription) {
            this.subscription.unsubscribe()
            this.subscription = null
        }
    }

    public unregisterAll(): void {
        this.selectors.clear()
        if (this.subscription) {
            this.subscription.unsubscribe()
            this.subscription = null
        }
    }
}

export class TextDocumentDidOpenFeature extends TextDocumentNotificationFeature<
    DidOpenTextDocumentParams,
    TextDocument
> {
    constructor(client: Client, environment: ObservableEnvironment) {
        super(
            client,
            environment.textDocument.pipe(filter((v): v is TextDocument => v !== null)),
            DidOpenTextDocumentNotification.type,
            client.options.middleware.didOpen,
            textDocument =>
                ({
                    textDocument: {
                        uri: textDocument.uri,
                        languageId: textDocument.languageId,
                        version: textDocument.version,
                        // TODO(sqs): add support for contents
                        text: 'getText' in textDocument ? textDocument.getText() : null,
                    },
                } as DidOpenTextDocumentParams),
            match
        )
    }

    public get messages(): typeof DidOpenTextDocumentNotification.type {
        return DidOpenTextDocumentNotification.type
    }

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        ensure(ensure(capabilities, 'textDocument')!, 'synchronization')!.dynamicRegistration = true
    }

    public initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector): void {
        if (
            documentSelector &&
            capabilities.textDocumentSync &&
            typeof capabilities.textDocumentSync !== 'number' &&
            capabilities.textDocumentSync.openClose
        ) {
            this.register(this.messages, {
                id: uuidv4(),
                registerOptions: { documentSelector },
            })
        }
    }
}
