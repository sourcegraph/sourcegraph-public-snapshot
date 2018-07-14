import { BehaviorSubject, from, Observable, TeardownLogic } from 'rxjs'
import * as uuidv4 from 'uuid/v4'
import { TextDocumentIdentifier } from 'vscode-languageserver-types'
import { ProvideTextDocumentDecorationsSignature } from '../../environment/providers/decoration'
import { TextDocumentFeatureProviderRegistry } from '../../environment/providers/textDocument'
import { ClientCapabilities, ServerCapabilities, TextDocumentRegistrationOptions } from '../../protocol'
import {
    TextDocumentDecoration,
    TextDocumentDecorationsParams,
    TextDocumentDecorationsRequest,
    TextDocumentPublishDecorationsNotification,
    TextDocumentPublishDecorationsParams,
} from '../../protocol/decoration'
import { DocumentSelector } from '../../types/document'
import { NextSignature } from '../../types/middleware'
import { Client } from '../client'
import { ensure, TextDocumentFeature } from './common'

export type ProvideTextDocumentDecorationsMiddleware = NextSignature<
    TextDocumentDecorationsParams,
    Observable<TextDocumentDecoration[] | null>
>

/**
 * Support for static text document decorations requested by the client (textDocument/decorations requests to the
 * server).
 */
export class TextDocumentStaticDecorationsFeature extends TextDocumentFeature<TextDocumentRegistrationOptions> {
    constructor(
        client: Client,
        private registry: TextDocumentFeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentDecorationsSignature
        >
    ) {
        super(client, TextDocumentDecorationsRequest.type)
    }

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        ensure(capabilities, 'decorations')!.static = true
    }

    public initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector): void {
        if (!capabilities.decorationsProvider || !capabilities.decorationsProvider.static || !documentSelector) {
            return
        }
        this.register(this.messages, {
            id: uuidv4(),
            registerOptions: { documentSelector },
        })
    }

    protected registerProvider(options: TextDocumentRegistrationOptions): TeardownLogic {
        const client = this.client
        const provideTextDocumentDecorations: ProvideTextDocumentDecorationsSignature = params =>
            from(client.sendRequest(TextDocumentDecorationsRequest.type, params))
        const middleware = client.clientOptions.middleware!
        return this.registry.registerProvider(
            options,
            (params: TextDocumentDecorationsParams): Observable<TextDocumentDecoration[] | null> =>
                middleware.provideTextDocumentDecorations
                    ? middleware.provideTextDocumentDecorations(params, provideTextDocumentDecorations)
                    : provideTextDocumentDecorations(params)
        )
    }
}

export type HandleTextDocumentDecorationsMiddleware = NextSignature<
    TextDocumentPublishDecorationsParams,
    Observable<TextDocumentDecoration[] | null>
>

/**
 * Support for dynamic text document decorations published by the server (textDocument/publishDecorations
 * notifications from the server).
 */
export class TextDocumentDynamicDecorationsFeature extends TextDocumentFeature<TextDocumentRegistrationOptions> {
    /** Map of document URI to its decorations (last published by the server). */
    private decorations = new Map<string, BehaviorSubject<TextDocumentDecoration[] | null>>()

    constructor(
        client: Client,
        private registry: TextDocumentFeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentDecorationsSignature
        >
    ) {
        super(client, TextDocumentPublishDecorationsNotification.type)
    }

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        ensure(capabilities, 'decorations')!.dynamic = true
    }

    public initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector): void {
        if (!capabilities.decorationsProvider || !capabilities.decorationsProvider.dynamic || !documentSelector) {
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

    protected registerProvider(options: TextDocumentRegistrationOptions): TeardownLogic {
        const client = this.client
        const provideTextDocumentDecorations: ProvideTextDocumentDecorationsSignature = params =>
            this.getDecorationsSubject(params.textDocument)
        const middleware = client.clientOptions.middleware!
        return this.registry.registerProvider(
            options,
            (params: TextDocumentDecorationsParams): Observable<TextDocumentDecoration[] | null> =>
                middleware.provideTextDocumentDecorations
                    ? middleware.provideTextDocumentDecorations(params, provideTextDocumentDecorations)
                    : provideTextDocumentDecorations(params)
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
