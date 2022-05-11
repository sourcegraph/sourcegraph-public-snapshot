import { isDefined, isErrorLike } from '@sourcegraph/common'

import { SettingsCascadeOrError } from '../settings/settings'

/**
 * Returns the allow only Sourcegraph authored extensions setting value from org or site settings.
 * Settings priority order: org, site, user. If it's not defined, returns `false`.
 */
export function allowOnlySourcegraphAuthoredExtensionsFromSettings(settingsCascade: SettingsCascadeOrError): boolean {
    if (!settingsCascade.subjects) {
        return false
    }

    const orgSubject = settingsCascade.subjects.find(({ subject }) => subject.__typename === 'Org')
    const siteSubject = settingsCascade.subjects.find(({ subject }) => subject.__typename === 'Site')
    const userSubject = settingsCascade.subjects.find(({ subject }) => subject.__typename === 'User')

    for (const subject of [orgSubject, siteSubject, userSubject].filter(isDefined)) {
        if (
            isErrorLike(subject.settings) ||
            subject.settings?.['extensions.allowOnlySourcegraphAuthored'] === undefined
        ) {
            continue
        }

        return subject.settings['extensions.allowOnlySourcegraphAuthored'] as boolean
    }

    return false
}
