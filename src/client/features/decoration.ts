import { BehaviorSubject, Observable, Unsubscribable } from 'rxjs'
import uuidv4 from 'uuid/v4'
import { TextDocumentIdentifier } from 'vscode-languageserver-types'
import { ProvideTextDocumentDecorationSignature } from '../../environment/providers/decoration'
import { FeatureProviderRegistry } from '../../environment/providers/registry'
import { ClientCapabilities, ServerCapabilities } from '../../protocol'
import { TextDocumentDecoration, TextDocumentPublishDecorationsNotification } from '../../protocol/decoration'
import { NextSignature } from '../../types/middleware'
import { Client } from '../client'
import { ensure, Feature } from './common'

export type ProvideTextDocumentDecorationMiddleware = NextSignature<
    TextDocumentIdentifier,
    Observable<TextDocumentDecoration[] | null>
>

/**
 * Support for text document decorations published by the server (textDocument/publishDecorations notifications
 * from the server).
 */
export class TextDocumentDecorationFeature extends Feature<undefined> {
    /** Map of document URI to its decorations (last published by the server). */
    private decorations = new Map<string, BehaviorSubject<TextDocumentDecoration[] | null>>()

    constructor(
        client: Client,
        private registry: FeatureProviderRegistry<undefined, ProvideTextDocumentDecorationSignature>
    ) {
        super(client)
    }

    public readonly messages = TextDocumentPublishDecorationsNotification.type

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        ensure(capabilities, 'decoration')
    }

    public initialize(capabilities: ServerCapabilities): void {
        if (!capabilities.decorationProvider) {
            return
        }
        this.register(this.messages, { id: uuidv4(), registerOptions: undefined })
        // TODO(sqs): no way to unregister this
        this.client.onNotification(TextDocumentPublishDecorationsNotification.type, params => {
            this.getDecorationsSubject(params.textDocument, params.decorations)
        })
    }

    protected registerProvider(): Unsubscribable {
        const client = this.client
        const provideTextDocumentDecoration: ProvideTextDocumentDecorationSignature = textDocument =>
            this.getDecorationsSubject(textDocument)
        const middleware = client.options.middleware.provideTextDocumentDecoration
        return this.registry.registerProvider(
            undefined,
            (textDocument: TextDocumentIdentifier): Observable<TextDocumentDecoration[] | null> =>
                middleware
                    ? middleware(textDocument, provideTextDocumentDecoration)
                    : provideTextDocumentDecoration(textDocument)
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
