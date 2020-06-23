import { Selection } from '@sourcegraph/extension-api-types'

export type Viewer = CodeEditor | DirectoryViewer
export type ViewerData = CodeEditorData | DirectoryViewerData

/**
 * ViewerId exposes the unique ID of a viewer.
 */
export interface ViewerId {
    /** The unique ID of the viewer. */
    readonly viewerId: string
}

export interface BaseViewerData {
    readonly isActive: boolean
}

export interface DirectoryViewerData extends BaseViewerData {
    readonly type: 'DirectoryViewer'
    /** The URI of the directory that this viewer is displaying. */
    readonly resource: string
}

/**
 * Describes a code viewer to be created.
 */
export interface CodeEditorData extends BaseViewerData {
    readonly type: 'CodeEditor'

    /** The URI of the model that this viewer is displaying. */
    readonly resource: string

    readonly selections: Selection[]
}

/**
 * Describes a code editor that has been added to the {@link ViewerService}.
 */
export interface CodeEditor extends ViewerId, CodeEditorData {}

/**
 * Describes a directory viewer that has been added to the {@link ViewerService}.
 */
export interface DirectoryViewer extends ViewerId, DirectoryViewerData {}

export type ViewerWithPartialModel = CodeEditorWithPartialModel | DirectoryViewer // Directories don't have a model

/**
 * A code editor with a partial model.
 *
 * To get the editor's full model, look up the model in the {@link ModelService}.
 */
export interface CodeEditorWithPartialModel extends CodeEditor {
    model: PartialModel
}

/**
 * A code editor with its full model, including the model text.
 */
export interface CodeEditorWithModel extends CodeEditor {
    /** The code editor's model. */
    model: TextModel
}

export type ViewerUpdate =
    | ({ type: 'added'; viewerData: ViewerData } & ViewerId)
    | ({ type: 'updated'; viewerData: Pick<CodeEditorData, 'selections'> } & ViewerId)
    | ({ type: 'deleted' } & ViewerId)
