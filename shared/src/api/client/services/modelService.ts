import { BehaviorSubject } from 'rxjs'
import { EMPTY_MODEL, Model } from '../model'
import { BehaviorSubjectLike } from './util'

/**
 * The model service manages the model of documents and roots.
 */
export interface ModelService {
    readonly model: BehaviorSubjectLike<Model>
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
