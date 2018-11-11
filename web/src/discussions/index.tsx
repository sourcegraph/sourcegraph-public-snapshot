import { isExtensionEnabled } from '../../../shared/src/extensions/extension'
import { Settings, SettingsCascadeOrError, SettingsSubject } from '../../../shared/src/settings'

/**
 * Tells whether or not the code discussions extensions is enabled or not.
 */
export function isDiscussionsEnabled(settingsCascade: SettingsCascadeOrError<SettingsSubject, Settings>): boolean {
    return isExtensionEnabled(settingsCascade.final, 'sourcegraph/code-discussions')
}
