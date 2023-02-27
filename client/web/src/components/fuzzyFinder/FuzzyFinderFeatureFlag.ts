import { SettingsExperimentalFeatures } from '@sourcegraph/shared/src/schema/settings.schema'
import { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

import { getExperimentalFeatures } from '../../util/get-experimental-features'

export function getFuzzyFinderFeatureFlags(
    finalSettings?: SettingsCascadeOrError<Settings>['final']
): Pick<
    SettingsExperimentalFeatures,
    'fuzzyFinderAll' | 'fuzzyFinderActions' | 'fuzzyFinderRepositories' | 'fuzzyFinderSymbols' | 'fuzzyFinderNavbar'
> {
    let { fuzzyFinderAll, fuzzyFinderActions, fuzzyFinderRepositories, fuzzyFinderSymbols, fuzzyFinderNavbar } =
        getExperimentalFeatures(finalSettings)
    // enable fuzzy finder unless it's explicitly disabled in settings
    fuzzyFinderAll = fuzzyFinderAll ?? true
    // Intentionally skip fuzzyFinderActions because we don't have enough actions implemented
    // Intentionally skip fuzzyFinderNavbar because the navbar is already too busy and we need to explore alternative solutions for the discoverability problem
    fuzzyFinderRepositories = fuzzyFinderAll || fuzzyFinderRepositories
    fuzzyFinderSymbols = fuzzyFinderAll || fuzzyFinderSymbols
    return { fuzzyFinderAll, fuzzyFinderActions, fuzzyFinderRepositories, fuzzyFinderSymbols, fuzzyFinderNavbar }
}
