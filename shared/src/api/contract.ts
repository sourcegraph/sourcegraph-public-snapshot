import { ProxyMarked } from 'comlink'
import { SettingsCascade } from '../settings/settings'
import { SettingsEdit } from './client/services/settings'

/**
 * APi for communicating in Main -> Ext Host direction
 */
export interface ExposedToClient extends ProxyMarked {
    // ExtConfiguration
    updateConfigurationData: (data: Readonly<SettingsCascade<object>>) => void
}

/**
 * APi for communicating in Ext Host -> Main direction
 */
export interface CalledFromExtHost extends ProxyMarked {
    // ExtConfiguration
    changeConfiguration: (edit: SettingsEdit) => void
}
