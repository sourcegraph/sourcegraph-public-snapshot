import { Subscribable, Subject } from 'rxjs'
import { TextDocument } from 'sourcegraph'

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
    /** All known models. */
    models: ReadonlyMap<string, TextModel>

    modelUpdates: Subscribable<readonly TextModelUpdate[]>

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
    const models = new Map<string, TextModel>()
    const modelUpdates = new Subject<TextModelUpdate[]>()
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
        addModel: model => {
            if (models.has(model.uri)) {
                throw new Error(`model already exists with URI ${model.uri}`)
            }
            models.set(model.uri, model)
            modelUpdates.next([{ type: 'added', ...model }])
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
            models.delete(uri)
            modelUpdates.next([{ type: 'deleted', uri }])
        },
    }
}
