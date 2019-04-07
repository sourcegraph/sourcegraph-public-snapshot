import { ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { flatten, values } from 'lodash'
import { BehaviorSubject, Observable, Subscription } from 'rxjs'
import { ProvideTextDocumentDecorationSignature } from '../services/decoration'
import { EditorService } from '../services/editorService'
import { FeatureProviderRegistry } from '../services/registry'
import { TextDocumentIdentifier } from '../types/textDocument'

/** @internal */
export interface ClientCodeEditorAPI extends ProxyValue {
    $setDecorations(resource: string, decorationType: string, decorations: TextDocumentDecoration[]): void
    $setCollapsed(editorId: string, collapsed: boolean): void
}

interface PreviousDecorations {
    [resource: string]: {
        [decorationType: string]: TextDocumentDecoration[]
    }
}

/** @internal */
export class ClientCodeEditor implements ClientCodeEditorAPI {
    public readonly [proxyValueSymbol] = true

    private subscriptions = new Subscription()

    /** Map of document URI to its decorations (last published by the server). */
    private decorations = new Map<string, BehaviorSubject<TextDocumentDecoration[]>>()

    private previousDecorations: PreviousDecorations = {}

    constructor(
        private registry: FeatureProviderRegistry<undefined, ProvideTextDocumentDecorationSignature>,
        private editorService: EditorService
    ) {
        this.subscriptions.add(
            this.registry.registerProvider(
                undefined,
                (textDocument: TextDocumentIdentifier): Observable<TextDocumentDecoration[]> =>
                    this.getDecorationsSubject(textDocument.uri)
            )
        )
    }

    public $setDecorations(resource: string, decorationType: string, decorations: TextDocumentDecoration[]): void {
        // tslint:disable-next-line: rxjs-no-ignored-observable
        this.getDecorationsSubject(resource, decorationType, decorations)
    }

    private getDecorationsSubject(
        resource: string,
        decorationType?: string,
        decorations?: TextDocumentDecoration[]
    ): BehaviorSubject<TextDocumentDecoration[]> {
        let subject = this.decorations.get(resource)
        if (!subject) {
            subject = new BehaviorSubject<TextDocumentDecoration[]>(decorations || [])
            this.decorations.set(resource, subject)
            this.previousDecorations[resource] = {}
        }
        if (decorations !== undefined) {
            // Replace previous decorations for this resource + decorationType
            this.previousDecorations[resource][decorationType!] = decorations

            // Merge decorations for all types for this resource, and emit them
            const nextDecorations = flatten(values(this.previousDecorations[resource]))
            subject.next(nextDecorations)
        }
        return subject
    }

    public $setCollapsed(editorId: string, collapsed: boolean): void {
        this.editorService.setCollapsed({ editorId }, collapsed)
    }

    public unsubscribe(): void {
        // Clear decorations.
        for (const subject of this.decorations.values()) {
            subject.next([])
        }

        this.subscriptions.unsubscribe()
    }
}
