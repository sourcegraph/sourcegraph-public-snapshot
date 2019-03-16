import { isExtensionEnabled } from '../../../shared/src/extensions/extension'
import { SettingsCascadeOrError } from '../../../shared/src/settings/settings'

/**
 * Tells whether or not the code discussions extensions is enabled or not.
 */
export function isDiscussionsEnabled(settingsCascade: SettingsCascadeOrError): boolean {
    return true // for use when sideloading
	/////////////// TODO!(sqs)
    return isExtensionEnabled(settingsCascade.final, 'sourcegraph/code-discussions')
}
