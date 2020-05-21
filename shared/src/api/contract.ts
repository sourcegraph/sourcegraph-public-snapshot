import { SettingsCascade } from '../settings/settings'
import { SettingsEdit } from './client/services/settings'

/**
 * This is exposed from the extension host thread to the main thread
 * e.g. for communicating  direction "main -> ext host"
 * Note this API object lives in the extension host thread
 */
export interface FlatExtHostAPI {
    // ExtConfiguration
    updateConfigurationData: (data: Readonly<SettingsCascade<object>>) => void
}

/**
 * This is exposed from the main thread to the extension host thread"
 * e.g. for communicating  direction "ext host -> main"
 * Note this API object lives in the main thread
 */
export interface MainThreadAPI {
    // ExtConfiguration
    changeConfiguration: (edit: SettingsEdit) => Promise<void>
}
