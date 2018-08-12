import { Observable, Subscription } from 'rxjs'
import { bufferCount, filter, map } from 'rxjs/operators'
import uuidv4 from 'uuid/v4'
import { TextDocument, TextDocumentItem } from 'vscode-languageserver-types'
import { MessageType as RPCMessageType, NotificationType } from '../../jsonrpc2/messages'
import {
    ClientCapabilities,
    DidCloseTextDocumentNotification,
    DidCloseTextDocumentParams,
    DidOpenTextDocumentNotification,
    DidOpenTextDocumentParams,
    ServerCapabilities,
    TextDocumentRegistrationOptions,
    TextDocumentSyncOptions,
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

/**
 * Support for notifying the server of when the client opened a text document (textDocument/didOpen notifications
 * sent by the client to the server).
 */
export class TextDocumentDidOpenFeature extends TextDocumentNotificationFeature<
    DidOpenTextDocumentParams,
    TextDocumentItem
> {
    constructor(client: Client, environmentTextDocument: Observable<Pick<TextDocument, 'uri' | 'languageId'> | null>) {
        super(
            client,
            environmentTextDocument.pipe(filter((v): v is TextDocumentItem => v !== null)),
            DidOpenTextDocumentNotification.type,
            client.options.middleware.didOpen,
            textDocument =>
                ({
                    textDocument,
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
        const textDocumentSyncOptions = resolveTextDocumentSync(capabilities.textDocumentSync)
        if (documentSelector && textDocumentSyncOptions && textDocumentSyncOptions.openClose) {
            this.register(this.messages, {
                id: uuidv4(),
                registerOptions: { documentSelector },
            })
        }
    }
}

/**
 * Support for notifying the server of when the client closed a text document (textDocument/didClose notifications
 * sent by the client to the server).
 */
export class TextDocumentDidCloseFeature extends TextDocumentNotificationFeature<
    DidCloseTextDocumentParams,
    TextDocumentItem
> {
    constructor(client: Client, environmentTextDocument: Observable<Pick<TextDocument, 'uri' | 'languageId'> | null>) {
        super(
            client,
            environmentTextDocument.pipe(
                bufferCount(2, 1),
                // When the previous value emitted was a document, it means the most recent value was either a
                // different document or null. In both cases, it means the previous document is closed.
                filter((v): v is [TextDocumentItem, TextDocumentItem | null] => v[0] !== null),
                map(([closedDocument]) => closedDocument)
            ),
            DidCloseTextDocumentNotification.type,
            client.options.middleware.didClose,
            textDocument =>
                ({
                    textDocument: { uri: textDocument.uri },
                } as DidCloseTextDocumentParams),
            match
        )
    }

    public get messages(): typeof DidCloseTextDocumentNotification.type {
        return DidCloseTextDocumentNotification.type
    }

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        ensure(ensure(capabilities, 'textDocument')!, 'synchronization')!.dynamicRegistration = true
    }

    public initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector): void {
        const textDocumentSyncOptions = resolveTextDocumentSync(capabilities.textDocumentSync)
        if (documentSelector && textDocumentSyncOptions && textDocumentSyncOptions.openClose) {
            this.register(this.messages, {
                id: uuidv4(),
                registerOptions: { documentSelector },
            })
        }
    }
}

export function resolveTextDocumentSync(
    value: ServerCapabilities['textDocumentSync']
): TextDocumentSyncOptions | undefined {
    if (!value) {
        return undefined
    } else if (typeof value === 'number') {
        return {
            openClose: true,
            change: value,
            save: { includeText: false },
        }
    }
    return value
}
