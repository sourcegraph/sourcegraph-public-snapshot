import React from 'react'
import { SettingsCascadeOrError } from '../settings/settings'
import { isErrorLike } from '../util/errors'

interface Props {
    settingsCascade: SettingsCascadeOrError
}

export const EmptyCommandList: React.FunctionComponent<Props> = ({ settingsCascade }) => {
    const onlyDefault = onlyDefaultExtensionsAdded(settingsCascade)

    return (
        <div>
            <p>No matching commands</p>
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
