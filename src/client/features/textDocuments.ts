import { Observable, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'
import * as uuidv4 from 'uuid/v4'
import { TextDocument } from 'vscode-languageserver-types'
import { MessageType as RPCMessageType, NotificationType } from '../../jsonrpc2/messages'
import {
    ClientCapabilities,
    DidOpenTextDocumentNotification,
    DidOpenTextDocumentParams,
    ServerCapabilities,
    TextDocumentRegistrationOptions,
} from '../../protocol'
import { DocumentSelector } from '../../types/documents'
import { NextSignature } from '../../types/middleware'
import { match } from '../../types/textDocument'
import { Client } from '../client'
import { DynamicFeature, ensure, RegistrationData } from './common'

type CreateParamsSignature<E, P> = (data: E) => P

export abstract class TextDocumentNotificationFeature<P, E> implements DynamicFeature<TextDocumentRegistrationOptions> {
    private subscriptions: Subscription | null = null

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
        this.selectors.set(data.id, data.registerOptions.documentSelector)
        if (!this.subscriptions) {
            this.subscriptions = this.observable.subscribe(data => this.callback(data))
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
        this.selectors.delete(id)
        if (this.selectors.size === 0 && this.subscriptions) {
            this.subscriptions.unsubscribe()
            this.subscriptions = null
        }
    }

    public unsubscribe(): void {
        this.selectors.clear()
        if (this.subscriptions) {
            this.subscriptions.unsubscribe()
            this.subscriptions = null
        }
    }
}

export class TextDocumentDidOpenFeature extends TextDocumentNotificationFeature<
    DidOpenTextDocumentParams,
    TextDocument
> {
    constructor(client: Client) {
        super(
            client,
            client.clientOptions.environment.textDocument.pipe(filter((v): v is TextDocument => v !== null)),
            DidOpenTextDocumentNotification.type,
            client.clientOptions.middleware!.didOpen,
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
