import { BehaviorSubject, NextObserver, Subscribable } from 'rxjs'
import { EMPTY_MODEL, Model } from '../model'

/**
 * The model service manages the model of documents and roots.
 */
export interface ModelService {
    readonly model: Subscribable<Model> & { readonly value: Model } & NextObserver<Model>
}

/**
 * Creates a {@link ModelService} instance.
 */
export function createModelService(): ModelService {
    const model = new BehaviorSubject<Model>(EMPTY_MODEL)
    return {
        model,
    }
}
