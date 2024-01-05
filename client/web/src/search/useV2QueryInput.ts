import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting, type UseTemporarySettingsReturnType } from '@sourcegraph/shared/src/settings/temporary'

/**
 * Whether or not to use the v2 search query input. Defaults to true if the experimental
 * feature flag is set. The user can override the value via temporary settings. While temporary
 * settings are loading the settings also defaults to true.
 */
export function useV2QueryInput(): [boolean, UseTemporarySettingsReturnType<'search.input.experimental'>[1]] {
    const v2QueryInputEnabled = useExperimentalFeatures(
        features =>
            features.searchQueryInput === 'v2' ||
            // support the old `experimental` name that refers to the same thing as `v2`
            (features.searchQueryInput as any) === 'experimental'
    )
    const [userSettingEnabled, setUserSetting] = useTemporarySetting('search.input.experimental', true)

    return [v2QueryInputEnabled && (userSettingEnabled ?? true), setUserSetting]
}
