import { BehaviorSubject, from, Observable, Unsubscribable } from 'rxjs'
import uuidv4 from 'uuid/v4'
import { TextDocumentIdentifier } from 'vscode-languageserver-types'
import { ProvideTextDocumentDecorationSignature } from '../../environment/providers/decoration'
import { TextDocumentFeatureProviderRegistry } from '../../environment/providers/textDocument'
import { ClientCapabilities, ServerCapabilities, TextDocumentRegistrationOptions } from '../../protocol'
import {
    TextDocumentDecoration,
    TextDocumentDecorationParams,
    TextDocumentDecorationRequest,
    TextDocumentPublishDecorationsNotification,
    TextDocumentPublishDecorationsParams,
} from '../../protocol/decoration'
import { DocumentSelector } from '../../types/document'
import { NextSignature } from '../../types/middleware'
import { Client } from '../client'
import { ensure, TextDocumentFeature } from './common'

export type ProvideTextDocumentDecorationMiddleware = NextSignature<
    TextDocumentDecorationParams,
    Observable<TextDocumentDecoration[] | null>
>

/**
 * Support for static text document decorations requested by the client (textDocument/decoration requests to the
 * server).
 */
export class TextDocumentStaticDecorationFeature extends TextDocumentFeature<TextDocumentRegistrationOptions> {
    constructor(
        client: Client,
        private registry: TextDocumentFeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentDecorationSignature
        >
    ) {
        super(client)
    }

    public readonly messages = TextDocumentDecorationRequest.type

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        ensure(capabilities, 'decoration')!.static = true
    }

    public initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector): void {
        if (!capabilities.decorationProvider || !capabilities.decorationProvider.static || !documentSelector) {
            return
        }
        this.register(this.messages, {
            id: uuidv4(),
            registerOptions: { documentSelector },
        })
    }

    protected registerProvider(options: TextDocumentRegistrationOptions): Unsubscribable {
        const client = this.client
        const provideTextDocumentDecoration: ProvideTextDocumentDecorationSignature = params =>
            from(client.sendRequest(TextDocumentDecorationRequest.type, params))
        const middleware = client.options.middleware.provideTextDocumentDecoration
        return this.registry.registerProvider(
            options,
            (params: TextDocumentDecorationParams): Observable<TextDocumentDecoration[] | null> =>
                middleware ? middleware(params, provideTextDocumentDecoration) : provideTextDocumentDecoration(params)
        )
    }
}

export type HandleTextDocumentDecorationMiddleware = NextSignature<
    TextDocumentPublishDecorationsParams,
    Observable<TextDocumentDecoration[] | null>
>

/**
 * Support for dynamic text document decorations published by the server (textDocument/publishDecorations
 * notifications from the server).
 */
export class TextDocumentDynamicDecorationFeature extends TextDocumentFeature<TextDocumentRegistrationOptions> {
    /** Map of document URI to its decorations (last published by the server). */
    private decorations = new Map<string, BehaviorSubject<TextDocumentDecoration[] | null>>()

    constructor(
        client: Client,
        private registry: TextDocumentFeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentDecorationSignature
        >
    ) {
        super(client)
    }

    public readonly messages = TextDocumentPublishDecorationsNotification.type

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        ensure(capabilities, 'decoration')!.dynamic = true
    }

    public initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector): void {
        if (!capabilities.decorationProvider || !capabilities.decorationProvider.dynamic || !documentSelector) {
            return
        }
        this.register(this.messages, {
            id: uuidv4(),
            registerOptions: { documentSelector },
        })
        // TODO(sqs): no way to unregister this
        this.client.onNotification(TextDocumentPublishDecorationsNotification.type, params => {
            this.getDecorationsSubject(params.textDocument, params.decorations)
        })
    }

    protected registerProvider(options: TextDocumentRegistrationOptions): Unsubscribable {
        const client = this.client
        const provideTextDocumentDecoration: ProvideTextDocumentDecorationSignature = params =>
            this.getDecorationsSubject(params.textDocument)
        const middleware = client.options.middleware.provideTextDocumentDecoration
        return this.registry.registerProvider(
            options,
            (params: TextDocumentDecorationParams): Observable<TextDocumentDecoration[] | null> =>
                middleware ? middleware(params, provideTextDocumentDecoration) : provideTextDocumentDecoration(params)
        )
    }

    private getDecorationsSubject(
        textDocument: TextDocumentIdentifier,
        value?: TextDocumentDecoration[] | null
    ): BehaviorSubject<TextDocumentDecoration[] | null> {
        let subject = this.decorations.get(textDocument.uri)
        if (!subject) {
            subject = new BehaviorSubject<TextDocumentDecoration[] | null>(value || null)
            this.decorations.set(textDocument.uri, subject)
        }
        if (value !== undefined) {
            subject.next(value)
        }
        return subject
    }
}
