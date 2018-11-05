import { BehaviorSubject, Observable, Subscription } from 'rxjs'
import { handleRequests } from '../../common/proxy'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { TextDocumentDecoration } from '../../protocol/plainTypes'
import { ProvideTextDocumentDecorationSignature } from '../providers/decoration'
import { FeatureProviderRegistry } from '../providers/registry'
import { TextDocumentIdentifier } from '../types/textDocument'

/** @internal */
export interface ClientCodeEditorAPI {
    $setDecorations(resource: string, decorations: TextDocumentDecoration[]): void
}

/** @internal */
export class ClientCodeEditor implements ClientCodeEditorAPI {
    private subscriptions = new Subscription()

    /** Map of document URI to its decorations (last published by the server). */
    private decorations = new Map<string, BehaviorSubject<TextDocumentDecoration[] | null>>()

    constructor(
        connection: Connection,
        private registry: FeatureProviderRegistry<undefined, ProvideTextDocumentDecorationSignature>
    ) {
        this.subscriptions.add(
            this.registry.registerProvider(
                undefined,
                (textDocument: TextDocumentIdentifier): Observable<TextDocumentDecoration[] | null> =>
                    this.getDecorationsSubject(textDocument.uri)
            )
        )

        handleRequests(connection, 'codeEditor', this)
    }

    public $setDecorations(resource: string, decorations: TextDocumentDecoration[]): void {
        this.getDecorationsSubject(resource, decorations)
    }

    private getDecorationsSubject(
        resource: string,
        value?: TextDocumentDecoration[] | null
    ): BehaviorSubject<TextDocumentDecoration[] | null> {
        let subject = this.decorations.get(resource)
        if (!subject) {
            subject = new BehaviorSubject<TextDocumentDecoration[] | null>(value || null)
            this.decorations.set(resource, subject)
        }
        if (value !== undefined) {
            subject.next(value)
        }
        return subject
    }

    public unsubscribe(): void {
        // Clear decorations.
        for (const subject of this.decorations.values()) {
            subject.next(null)
        }

        this.subscriptions.unsubscribe()
    }
}
