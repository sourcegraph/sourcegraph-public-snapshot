import React from 'react'
import { SettingsCascadeOrError } from '../settings/settings'
import { isErrorLike } from '../util/errors'

interface Props {
    settingsCascade?: SettingsCascadeOrError
    sourcegraphURL: string
}

export const EmptyCommandList: React.FunctionComponent<Props> = ({ settingsCascade, sourcegraphURL }) => {
    // if no settings cascade (yet), default to 'no active extensions'
    const onlyDefault = settingsCascade ? onlyDefaultExtensionsAdded(settingsCascade) : false
    /**
     * Three questions:
     * - no extensions enabled at all? or added on top of default?
     * - how to deal with styling?
     * - On the web app, should I use a react router link?
     */
    return (
        <div>
            <p>{onlyDefault ? "You don't have any extensions enabled" : "You don't have any active actions"}</p>
            <p>
                {onlyDefault
                    ? 'Find extensions in the Sourcegraph extension registry, or learn how to write your own in just a few minutes.'
                    : 'Commands from your installed extensions will be shown when you navigate to certain pages'}
            </p>
            <a href={sourcegraphURL + '/extensions'}>Explore extensions</a>
            <p style={{ position: 'relative', bottom: 0, right: 0 }}>icon</p>
        </div>
    )
}

/**
 * Determines if only default extensions are added
 */
function onlyDefaultExtensionsAdded(settings: SettingsCascadeOrError): boolean {
    if (!isErrorLike(settings.subjects) && settings.subjects) {
        const userSettings = settings.subjects.find(subject => subject.subject.__typename === 'User')
        const defaultSettings = settings.subjects.find(subject => subject.subject.__typename === 'DefaultSettings')

        if (userSettings && defaultSettings) {
            const userExtensions = !isErrorLike(userSettings.settings) && userSettings.settings?.extensions
            const defaultExtensions = !isErrorLike(defaultSettings.settings) && defaultSettings.settings?.extensions

            if (userExtensions && defaultExtensions) {
                for (const key of Object.keys(userExtensions)) {
                    if (!(key in defaultExtensions)) {
                        return false
                    }
                }
            }
        }
    }

    return true
}
