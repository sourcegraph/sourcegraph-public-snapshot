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
    // fuzzyFinderAll = true // Uncomment this line if working with `sg start web-standalone`
    fuzzyFinderActions = fuzzyFinderAll || fuzzyFinderActions
    fuzzyFinderRepositories = fuzzyFinderAll || fuzzyFinderRepositories
    fuzzyFinderNavbar = fuzzyFinderAll || fuzzyFinderNavbar
    fuzzyFinderSymbols = fuzzyFinderAll || fuzzyFinderSymbols
    return { fuzzyFinderAll, fuzzyFinderActions, fuzzyFinderRepositories, fuzzyFinderSymbols, fuzzyFinderNavbar }
}
