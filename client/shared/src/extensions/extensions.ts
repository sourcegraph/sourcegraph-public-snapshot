import { Settings, SettingsCascadeOrError, SettingsSubject } from '../settings/settings'
import { isErrorLike } from '../util/errors'

/**
 * Determines if only default extensions are added
 */
export function onlyDefaultExtensionsAdded(settings: SettingsCascadeOrError): boolean {
    const userExtensions = getEnabledExtensionsForSubject(settings, 'User')
    const defaultExtensions = getEnabledExtensionsForSubject(settings, 'DefaultSettings')

    if (userExtensions && defaultExtensions) {
        for (const key of Object.keys(userExtensions)) {
            if (!(key in defaultExtensions)) {
                return false
            }
        }
    }

    return true
}

export function getEnabledExtensionsForSubject(
    settings: SettingsCascadeOrError,
    settingsSubject: SettingsSubject['__typename']
): Settings['extensions'] {
    if (!isErrorLike(settings.subjects) && settings.subjects) {
        const defaultSettings = settings.subjects.find(subject => subject.subject.__typename === settingsSubject)

        if (defaultSettings && !isErrorLike(defaultSettings.settings)) {
            return defaultSettings.settings?.extensions
        }
    }

    return undefined
}

export function areExtensionsSame(oldExtensions: { id: string }[], newExtensions: { id: string }[]): boolean {
    const oldSet = new Set<string>(oldExtensions.map(({ id }) => id))
    const newSet = new Set<string>(newExtensions.map(({ id }) => id))

    for (const oldId of oldSet) {
        if (!newSet.has(oldId)) {
            return false
        }
    }

    for (const newId of newSet) {
        if (!oldSet.has(newId)) {
            return false
        }
    }

    return true
}
