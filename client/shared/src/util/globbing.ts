import { isErrorLike } from '@sourcegraph/common'

import { SettingsCascadeOrError } from '../settings/settings'

/**
 * Returns "true" if search.globbing is set to true in the final settings, "false" otherwise
 */
export const globbingEnabledFromSettings = (settings: SettingsCascadeOrError): boolean =>
    !!(settings.final && !isErrorLike(settings.final) && settings.final['search.globbing'])
