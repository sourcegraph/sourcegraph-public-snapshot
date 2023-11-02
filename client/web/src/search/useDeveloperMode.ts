import { useTemporarySetting, type UseTemporarySettingsReturnType } from '@sourcegraph/shared/src/settings/temporary'

export function useDeveloperMode(): [boolean, UseTemporarySettingsReturnType<'dev.developerMode.enabled'>[1]] {
    const [userSettingEnabled, setUserSetting] = useTemporarySetting('dev.developerMode.enabled', false)
    return [userSettingEnabled ?? false, setUserSetting]
}
