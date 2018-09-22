import { isExtensionEnabled } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import {
    ConfigurationCascadeOrError,
    ConfigurationSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'

/**
 * Tells whether or not the code discussions extensions is enabled or not.
 */
export function isDiscussionsEnabled(
    configurationCascade: ConfigurationCascadeOrError<ConfigurationSubject, Settings>
): boolean {
    return isExtensionEnabled(configurationCascade.merged, 'sourcegraph/code-discussions')
}
