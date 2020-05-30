import { SettingsCascade } from '../settings/settings'
import { SettingsEdit } from './client/services/settings'
import * as clientType from '@sourcegraph/extension-api-types'

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

    syncRoots(roots: readonly clientType.WorkspaceRoot[]): void
    syncVersionContext(versionContext: string | undefined): void
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
}
