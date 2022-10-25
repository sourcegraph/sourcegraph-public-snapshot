import { SettingsExperimentalFeatures } from '@sourcegraph/shared/src/schema/settings.schema'
import { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

import { getExperimentalFeatures } from '../../util/get-experimental-features'

export function getFuzzyFinderFeatureFlags(
    finalSettings?: SettingsCascadeOrError<Settings>['final']
): Pick<
    SettingsExperimentalFeatures,
    'fuzzyFinderAll' | 'fuzzyFinderActions' | 'fuzzyFinderRepositories' | 'fuzzyFinderSymbols' | 'fuzzyFinderNavbar'
> {
    let {
        fuzzyFinderAll,
        fuzzyFinderActions,
        fuzzyFinderRepositories,
        fuzzyFinderSymbols,
        fuzzyFinderNavbar,
    } = getExperimentalFeatures(finalSettings)
    // Intentionally skip "Actions" when `fuzzyFinderAll` is true
    fuzzyFinderRepositories = fuzzyFinderAll || fuzzyFinderRepositories
    fuzzyFinderNavbar = fuzzyFinderAll || fuzzyFinderNavbar
    fuzzyFinderSymbols = fuzzyFinderAll || fuzzyFinderSymbols
    return { fuzzyFinderAll, fuzzyFinderActions, fuzzyFinderRepositories, fuzzyFinderSymbols, fuzzyFinderNavbar }
}
