import { isErrorLike } from '@sourcegraph/common'

import { SettingsCascadeOrError } from '../settings/settings'

/**
 * Returns the allow only Sourcegraph authored extensions setting value from org or site settings.
 * Settings priority in descending order: site, org, user. If it's not defined, returns `false`.
 */
export function allowOnlySourcegraphAuthoredExtensionsFromSettings(settingsCascade: SettingsCascadeOrError): boolean {
    for (const type of ['Site', 'Org', 'User']) {
        const subject = settingsCascade.subjects?.find(({ subject }) => subject.__typename === type)

        if (
            !subject ||
            isErrorLike(subject.settings) ||
            subject.settings?.['extensions.allowOnlySourcegraphAuthored'] === undefined
        ) {
            continue
        }

        return subject.settings['extensions.allowOnlySourcegraphAuthored'] as boolean
    }

    return false
}
