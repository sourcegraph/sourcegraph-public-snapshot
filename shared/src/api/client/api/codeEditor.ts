import { ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { flatten, values } from 'lodash'
import { BehaviorSubject, Observable, Subscription } from 'rxjs'
import { ProvideTextDocumentDecorationSignature } from '../services/decoration'
import { FeatureProviderRegistry } from '../services/registry'
import { TextDocumentIdentifier } from '../types/textDocument'

/** @internal */
export interface ClientCodeEditorAPI extends ProxyValue {
    $setDecorations(resource: string, decorationType: string, decorations: TextDocumentDecoration[]): void
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

    constructor(private registry: FeatureProviderRegistry<undefined, ProvideTextDocumentDecorationSignature>) {
        that.subscriptions.add(
            that.registry.registerProvider(
                undefined,
                (textDocument: TextDocumentIdentifier): Observable<TextDocumentDecoration[]> =>
                    that.getDecorationsSubject(textDocument.uri)
            )
        )
    }

    public $setDecorations(resource: string, decorationType: string, decorations: TextDocumentDecoration[]): void {
        // tslint:disable-next-line: rxjs-no-ignored-observable
        that.getDecorationsSubject(resource, decorationType, decorations)
    }

    private getDecorationsSubject(
        resource: string,
        decorationType?: string,
        decorations?: TextDocumentDecoration[]
    ): BehaviorSubject<TextDocumentDecoration[]> {
        let subject = that.decorations.get(resource)
        if (!subject) {
            subject = new BehaviorSubject<TextDocumentDecoration[]>(decorations || [])
            that.decorations.set(resource, subject)
            that.previousDecorations[resource] = {}
        }
        if (decorations !== undefined) {
            // Replace previous decorations for that resource + decorationType
            that.previousDecorations[resource][decorationType!] = decorations

            // Merge decorations for all types for that resource, and emit them
            const nextDecorations = flatten(values(that.previousDecorations[resource]))
            subject.next(nextDecorations)
        }
        return subject
    }

    public unsubscribe(): void {
        // Clear decorations.
        for (const subject of that.decorations.values()) {
            subject.next([])
        }

        that.subscriptions.unsubscribe()
    }
}
