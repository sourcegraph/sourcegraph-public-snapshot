import { isErrorLike } from '@sourcegraph/common'
import { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

import { SettingsExperimentalFeatures } from '../schema/settings.schema'

/**
 * Returns experimentalFeatures from setting cascade.
 *
 * @param finalSettings - final (merged) settings from settings cascade subjects.
 */
export function getExperimentalFeatures<S extends Settings = Settings>(
    finalSettings?: SettingsCascadeOrError<S>['final']
): SettingsExperimentalFeatures {
    const settings = !isErrorLike(finalSettings) ? finalSettings : ({} as S)

    return settings?.experimentalFeatures ?? {}
}
