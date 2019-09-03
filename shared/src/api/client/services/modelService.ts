import { Subscribable, Subject, BehaviorSubject } from 'rxjs'
import { TextDocument } from 'sourcegraph'
import { RefCount } from '../../../util/RefCount'

/**
 * A text model is a text document and associated metadata.
 *
 * How does this relate to editors (in {@link EditorService}? A model is the file, an editor is the
 * window that the file is shown in. Things like the content and language are properties of the
 * model; things like decorations and the selection ranges are properties of the editor.
 */
export interface TextModel extends Pick<TextDocument, 'uri' | 'languageId' | 'text'> {}

export type TextModelUpdate =
    | { type: 'added' } & TextModel
    | { type: 'updated'; text: string } & Pick<TextModel, 'uri'>
    | { type: 'deleted' } & Pick<TextModel, 'uri'>

/**
 * The model service manages document contents and metadata.
 *
 * See {@link Model} for an explanation of the difference between a model and an editor.
 */
export interface ModelService {
    /**
     * A map of all known models, indexed by URI.
     */
    models: ReadonlyMap<string, TextModel>

    /**
     * An observable of all model updates.
     *
     * Emits when a model is added, updated or removed.
     */
    modelUpdates: Subscribable<readonly TextModelUpdate[]>

    /**
     * An observable of unique languageIds across all models.
     */
    activeLanguages: Subscribable<readonly string[]>

    /**
     * Adds a model.
     *
     * @param model The model to add.
     */
    addModel(model: TextModel): void

    /**
     * Updates a model's text content.
     *
     * @param uri The URI of the model whose content to update.
     * @param text The new text content (which will overwrite the model's previous content).
     * @throws if the model does not exist.
     */
    updateModel(uri: string, text: string): void

    /**
     * Reports whether a model with a given URI has already been added.
     *
     * @param uri The model URI to check.
     */
    hasModel(uri: string): boolean

    /**
     * Removes a model
     */
    removeModel(uri: string): void

    /**
     * Adds a reference from an editor to the model with the given URI.
     */
    addModelRef(uri: string): void

    /**
     * Removes a reference from an editor to the model with the given URI.
     * A model with zero references will be removed.
     */
    removeModelRef(uri: string): void
}

/**
 * Creates a new instance of {@link ModelService}.
 */
export function createModelService(): ModelService {
    /** A map of URIs to TextModels */
    const models = new Map<string, TextModel>()
    const modelUpdates = new Subject<TextModelUpdate[]>()
    const activeLanguages = new BehaviorSubject<string[]>([])
    const modelRefs = new RefCount()
    const languageRefs = new RefCount()
    const getModel = (uri: string): TextModel => {
        const model = models.get(uri)
        if (!model) {
            throw new Error(`model does not exist with URI ${uri}`)
        }
        return model
    }
    return {
        models,
        modelUpdates,
        activeLanguages,
        addModel: model => {
            if (models.has(model.uri)) {
                throw new Error(`model already exists with URI ${model.uri}`)
            }
            models.set(model.uri, model)
            modelUpdates.next([{ type: 'added', ...model }])
            if (languageRefs.increment(model.languageId)) {
                activeLanguages.next([...languageRefs.keys()])
            }
        },
        updateModel: (uri, text) => {
            const model = getModel(uri)
            models.set(uri, {
                ...model,
                text,
            })
            modelUpdates.next([{ type: 'updated', uri, text }])
        },
        hasModel: uri => models.has(uri),
        removeModel: uri => {
            const model = models.get(uri)
            if (!model) {
                throw new Error(`removeModel(): model not found ${uri}`)
            }
            models.delete(uri)
            modelRefs.delete(uri)
            modelUpdates.next([{ type: 'deleted', uri }])
            if (languageRefs.decrement(model.languageId)) {
                activeLanguages.next([...languageRefs.keys()])
            }
        },
        addModelRef: uri => {
            modelRefs.increment(uri)
        },
        removeModelRef: uri => {
            console.log('removeModelRef', uri)
            const model = models.get(uri)
            if (!model) {
                throw new Error(`removeModelRef(): model not found ${uri}`)
            }
            if (modelRefs.decrement(uri)) {
                models.delete(uri)
                modelUpdates.next([{ type: 'deleted', uri }])
                if (languageRefs.decrement(model.languageId)) {
                    activeLanguages.next([...languageRefs.keys()])
                }
            }
        },
    }
}
