import { isErrorLike } from '@sourcegraph/common'
import { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

export type SettingsGetter = (setting: string, defaultValue: boolean) => boolean

export function newSettingsGetter(settingsCascade: SettingsCascadeOrError<Settings>): SettingsGetter {
    return (setting: string, defaultValue: boolean) =>
        (settingsCascade.final && !isErrorLike(settingsCascade.final) && (settingsCascade.final[setting] as boolean)) ??
        defaultValue
}
