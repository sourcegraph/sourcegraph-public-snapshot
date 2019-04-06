import { BehaviorSubject, NextObserver, Subscribable } from 'rxjs'
import { EMPTY_MODEL, Model } from '../model'

/**
 * The editor service manages the model of documents and roots.
 */
export interface EditorService {
    readonly model: Subscribable<Model> & { readonly value: Model } & NextObserver<Model>
}

/**
 * Creates a {@link EditorService} instance.
 */
export function createEditorService(): EditorService {
    const model = new BehaviorSubject<Model>(EMPTY_MODEL)
    return {
        model,
    }
}
