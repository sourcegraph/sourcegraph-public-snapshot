import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting, UseTemporarySettingsReturnType } from '@sourcegraph/shared/src/settings/temporary'

/**
 * Whether or not to use the experimental search input. Defaults to true if the experimental
 * feature flag is set. The user can override the value via temporary settings. While temporary
 * settings are loading the settings also defaults to true.
 */
export function useExperimentalQueryInput(): [boolean, UseTemporarySettingsReturnType<'search.input.experimental'>[1]] {
    const experimentalSearchInputEnabled = useExperimentalFeatures(
        features => features.searchQueryInput === 'experimental'
    )
    const [userSettingEnabled, setUserSetting] = useTemporarySetting('search.input.experimental', true)

    return [experimentalSearchInputEnabled && (userSettingEnabled ?? true), setUserSetting]
}
