import { BehaviorSubject, Subscribable } from 'rxjs'
import { TextDocument } from 'sourcegraph'

/**
 * A text model is a text document and associated metadata.
 *
 * How does this relate to editors (in {@link EditorService}? A model is the file, an editor is the
 * window that the file is shown in. Things like the content and language are properties of the
 * model; things like decorations and the selection ranges are properties of the editor.
 */
export interface TextModel extends Pick<TextDocument, 'uri' | 'languageId' | 'text'> {}

/**
 * The model service manages document contents and metadata.
 *
 * See {@link Model} for an explanation of the difference between a model and an editor.
 */
export interface ModelService {
    /** All known models. */
    models: Subscribable<readonly TextModel[]>

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
     * Removes a model.
     *
     * @param uri The URI of the model to remove.
     */
    removeModel(uri: string): void
}

/**
 * Creates a new instance of {@link ModelService}.
 */
export function createModelService(): ModelService {
    /** A map of URIs to TextModels */
    const modelMap = new Map<string, TextModel>()
    const models = new BehaviorSubject<readonly TextModel[]>([])
    const hasModel = (uri: string): boolean => modelMap.has(uri)
    return {
        models,
        addModel: model => {
            if (hasModel(model.uri)) {
                throw new Error(`model already exists with URI ${model.uri}`)
            }
            modelMap.set(model.uri, model)
            models.next([...modelMap.values()])
        },
        updateModel: (uri, text) => {
            const existing = modelMap.get(uri)
            if (!existing) {
                throw new Error(`model does not exist with URI ${uri}`)
            }
            modelMap.set(uri, {
                ...existing,
                text,
            })
            models.next([...modelMap.values()])
        },
        hasModel,
        removeModel: uri => {
            modelMap.delete(uri)
            models.next([...modelMap.values()])
        },
    }
}
