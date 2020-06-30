import { SettingsCascade } from '../settings/settings'
import { SettingsEdit } from './client/services/settings'
import * as clientType from '@sourcegraph/extension-api-types'
import { Remote, ProxyMarked } from 'comlink'
import { Unsubscribable, TextDocument } from 'sourcegraph'
import { ProxySubscribable } from './extension/api/common'
import { TextDocumentPositionParams } from './protocol'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { HoverMerged } from './client/types/hover'
import { ViewerData, ViewerId } from './viewerTypes'
import { ConfiguredExtension } from '../extensions/extension'

/**
 * A text model is a text document and associated metadata.
 *
 * How does this relate to editors (in {@link ViewerService}? A model is the file, an editor is the
 * window that the file is shown in. Things like the content and language are properties of the
 * model; things like decorations and the selection ranges are properties of the editor.
 */
export interface TextModel extends Pick<TextDocument, 'uri' | 'languageId' | 'text'> {}

/**
 * This is exposed from the extension host thread to the main thread
 * e.g. for communicating  direction "main -> ext host"
 * Note this API object lives in the extension host thread
 */
export interface FlatExtHostAPI {
    /**
     * Updates the settings exposed to extensions.
     */
    syncSettingsData: (data: Readonly<SettingsCascade<object>>) => void

    // Workspace
    syncRoots: (roots: readonly clientType.WorkspaceRoot[]) => void
    syncVersionContext: (versionContext: string | undefined) => void

    // Search
    transformSearchQuery: (query: string) => ProxySubscribable<string>

    // Documents
    addTextDocumentIfNotExists(model: TextModel): Pick<TextDocument, 'uri'>

    // Viewers
    addViewerIfNotExists(model: ViewerData): ViewerId

    /**
     * Sets the selections for a CodeEditor.
     *
     * @param codeEditor The editor for which to set the selections.
     * @param selections The new selections to apply.
     * @throws if no editor exists with the given editor ID.
     * @throws if the editor ID is not a CodeEditor.
     */
    setEditorSelections(codeEditor: ViewerId, selections: clientType.Selection[]): void

    /**
     * Removes a viewer.
     * Also removes the corresponding model if no other viewer is referencing it.
     *
     * @param viewer The viewer to remove.
     */
    removeViewer(viewer: ViewerId): void

    // Languages
    getHover(parameters: TextDocumentPositionParams): ProxySubscribable<MaybeLoadingResult<HoverMerged | null>>
    getDefinitions(parameters: TextDocumentPositionParams): ProxySubscribable<MaybeLoadingResult<clientType.Location[]>>
    getReferences(parameters: TextDocumentPositionParams): ProxySubscribable<MaybeLoadingResult<clientType.Location[]>>
}

/**
 * This is exposed from the main thread to the extension host thread"
 * e.g. for communicating  direction "ext host -> main"
 * Note this API object lives in the main thread
 */
export interface MainThreadAPI {
    /**
     * Applies a settings update from extensions.
     */
    applySettingsEdit: (edit: SettingsEdit) => Promise<void>

    // Commands
    executeCommand: (command: string, args: any[]) => Promise<any>
    registerCommand: (
        name: string,
        command: Remote<((...args: any) => any) & ProxyMarked>
    ) => Unsubscribable & ProxyMarked

    getActiveExtensions: () => ProxySubscribable<ConfiguredExtension[]>
    getScriptURLForExtension: (bundleURL: string) => Promise<string>
}
