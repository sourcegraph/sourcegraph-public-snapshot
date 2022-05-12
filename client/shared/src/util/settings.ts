import { isErrorLike } from '@sourcegraph/common'

import { SettingsSubject } from '../schema'
import { SettingsCascadeOrError } from '../settings/settings'

/**
 * Returns the allow only Sourcegraph authored extensions setting value from org or site settings.
 * Settings priority in descending order: site, org, user. If it's not defined, returns `false`.
 */
export function allowOnlySourcegraphAuthoredExtensionsFromSettings(
    settingsCascade: SettingsCascadeOrError
): { value: boolean; subject?: SettingsSubject['__typename'] } {
    const types: SettingsSubject['__typename'][] = ['Site', 'Org', 'User']

    for (const type of types) {
        const subject = settingsCascade.subjects?.find(({ subject }) => subject.__typename === type)

        if (
            !subject ||
            isErrorLike(subject.settings) ||
            subject.settings?.['extensions.allowOnlySourcegraphAuthored'] === undefined
        ) {
            continue
        }

        const value = subject.settings['extensions.allowOnlySourcegraphAuthored'] as boolean

        return { value, subject: value ? type : undefined }
    }

    return { value: false }
}
