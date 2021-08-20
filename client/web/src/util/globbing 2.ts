import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

/**
 * Returns "true" if search.globbing is set to true in the final settings, "false" otherwise
 */
export const globbingEnabledFromSettings = (settings: SettingsCascadeOrError): boolean =>
    !!(settings.final && !isErrorLike(settings.final) && settings.final['search.globbing'])
