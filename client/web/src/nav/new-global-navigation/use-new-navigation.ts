import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting, type UseTemporarySettingsReturnType } from '@sourcegraph/shared/src/settings/temporary'

/**
 * Whether or not to use the v2 search navigation UI. Defaults to true if the experimental
 * feature flag is set. The user can override the value via temporary settings. While temporary
 * settings are loading the settings also defaults to true.
 */
export function useNewSearchNavigation(): [boolean, UseTemporarySettingsReturnType<'search.navigation'>[1]] {
    const newSearchNavigationUI = useExperimentalFeatures(features => features.newSearchNavigationUI ?? false)
    const [userSettingEnabled, setUserSetting] = useTemporarySetting('search.navigation', false)

    if (newSearchNavigationUI) {
        return [true, setUserSetting]
    }

    return [userSettingEnabled ?? true, setUserSetting]
}
