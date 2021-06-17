import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { SettingsExperimentalFeatures } from '../../schema/settings.schema'

/**
 * Code insights display location setting to check setting for particular view
 * to show code insights components.
 */
interface CodeInsightsDisplayLocation {
    insightsPage: boolean
    homepage: boolean
    directory: boolean
}

/**
 * Feature guard for code insights.
 *
 * @param settingsCascade - settings cascade object
 * @param views - Map with display location of insights {@link CodeInsightsDisplayLocation}
 */
export function isCodeInsightsEnabled(
    settingsCascade: SettingsCascadeOrError,
    views: Partial<CodeInsightsDisplayLocation> = {}
): boolean {
    if (isErrorLike(settingsCascade.final)) {
        return false
    }

    const final = settingsCascade.final
    const viewsKeys = Object.keys(views) as (keyof CodeInsightsDisplayLocation)[]
    const experimentalFeatures: SettingsExperimentalFeatures = final?.experimentalFeatures ?? {}

    if (!experimentalFeatures.codeInsights) {
        return false
    }

    return viewsKeys.every(viewKey => {
        if (views[viewKey]) {
            return final?.[`insights.displayLocation.${viewKey}`] !== false
        }

        return true
    })
}
