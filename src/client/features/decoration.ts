import { BehaviorSubject, Observable } from 'rxjs'
import { ProvideTextDocumentDecorationSignature } from '../../environment/providers/decoration'
import { FeatureProviderRegistry } from '../../environment/providers/registry'
import { ClientCapabilities } from '../../protocol'
import { TextDocumentDecoration, TextDocumentPublishDecorationsNotification } from '../../protocol/decoration'
import { TextDocumentIdentifier } from '../../types/textDocument'
import { Client } from '../client'
import { ensure, StaticFeature } from './common'

/**
 * Support for text document decorations published by the server (textDocument/publishDecorations notifications
 * from the server).
 */
export class TextDocumentDecorationFeature implements StaticFeature {
    /** Map of document URI to its decorations (last published by the server). */
    private decorations = new Map<string, BehaviorSubject<TextDocumentDecoration[] | null>>()

    constructor(
        private client: Client,
        private registry: FeatureProviderRegistry<undefined, ProvideTextDocumentDecorationSignature>
    ) {
        this.registry.registerProvider(
            undefined,
            (textDocument: TextDocumentIdentifier): Observable<TextDocumentDecoration[] | null> =>
                this.getDecorationsSubject(textDocument)
        )
    }

    public readonly messages = TextDocumentPublishDecorationsNotification.type

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        ensure(capabilities, 'decoration')
    }

    public initialize(): void {
        // TODO(sqs): no way to unregister this
        this.client.onNotification(TextDocumentPublishDecorationsNotification.type, params => {
            this.getDecorationsSubject(params.textDocument, params.decorations)
        })
    }

    public deinitialize(): void {
        // Clear decorations;
        for (const subject of Object.values(this.decorations)) {
            subject.next(null)
        }
    }

    protected validateRegistrationOptions(_data: any): undefined {
        return
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
