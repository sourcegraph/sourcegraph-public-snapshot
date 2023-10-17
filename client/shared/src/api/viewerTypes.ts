import type { Selection } from '@sourcegraph/extension-api-types'

import type { TextDocument } from '../codeintel/legacy-extensions/api'

import type { ExtensionCodeEditor } from './extension/api/codeEditor'
import type { ExtensionDirectoryViewer } from './extension/api/directoryViewer'

export type ExtensionViewer = ExtensionCodeEditor | ExtensionDirectoryViewer

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

export interface CodeEditor extends ViewerId, CodeEditorData {}

/**
 * Describes a directory viewer that has been added to the {@link ViewerService}.
 */
export interface DirectoryViewer extends ViewerId, DirectoryViewerData {}

export type ViewerWithPartialModel = CodeEditorWithPartialModel | DirectoryViewer // Directories don't have a model

/**
 * A code editor with a partial model.
 */
export interface CodeEditorWithPartialModel extends CodeEditor {
    model: PartialModel
}

/**
 * A text model is a text document and associated metadata.
 *
 * How does this relate to editors? A text document is the file, an editor is the
 * window that the file is shown in. Things like the content and language are properties of the
 * text document; things like selection ranges are properties of the editor.
 */
export interface TextDocumentData extends Pick<TextDocument, 'uri' | 'languageId' | 'text'> {}

/**
 * A partial {@link TextDocumentData}, containing only the fields that are
 * guaranteed to never be updated.
 */
export interface PartialModel extends Pick<TextDocumentData, 'languageId'> {}

export type ViewerUpdate = ViewerId & { action: 'removal' | 'addition'; type: 'DirectoryViewer' | 'CodeEditor' }
