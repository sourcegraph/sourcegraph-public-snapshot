import { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { SettingsExperimentalFeatures } from '../../schema/settings.schema'

/**
 * Returns experimentalFeatures from setting cascade.
 *
 * @param settingsCascade - Possible settings cascade object or error.
 */
export function getExperimentalFeatures<S extends Settings = Settings>(
    settingsCascade: SettingsCascadeOrError<S>
): SettingsExperimentalFeatures {
    const settings = !isErrorLike(settingsCascade.final) ? settingsCascade.final : ({} as S)

    return settings?.experimentalFeatures ?? {}
}
