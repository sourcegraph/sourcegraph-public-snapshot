import { isErrorLike } from '@sourcegraph/common'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import * as GQL from '@sourcegraph/shared/src/schema'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

import { AuthenticatedUser } from '../auth'
import { LayoutProps } from '../Layout'

/** A fallback settings subject that can be constructed synchronously at initialization time. */
export function siteSubjectNoAdmin(): Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'> {
    return {
        id: window.context.siteGQLID,
        viewerCanAdminister: false,
    }
}

export function viewerSubjectFromSettings(
    cascade: SettingsCascadeOrError,
    authenticatedUser?: AuthenticatedUser | null
): LayoutProps['viewerSubject'] {
    if (authenticatedUser) {
        return authenticatedUser
    }
    if (cascade && !isErrorLike(cascade) && cascade.subjects && cascade.subjects.length > 0) {
        return cascade.subjects[0].subject
    }
    return siteSubjectNoAdmin()
}

/**
 * Returns the user-configured search pattern type or undefined if not
 * configured by the user.
 */
export function defaultPatternTypeFromSettings(settingsCascade: SettingsCascadeOrError): SearchPatternType | undefined {
    return getFromSettings(settingsCascade, 'search.defaultPatternType')
}

/**
 * Returns the user-configured case sensitivity setting or undefined if not
 * configured by the user.
 */
export function defaultCaseSensitiveFromSettings(settingsCascade: SettingsCascadeOrError): boolean | undefined {
    return getFromSettings(settingsCascade, 'search.defaultCaseSensitive')
}

/**
 * Whether user wants to use default extensions functionality as core product features instead of the corresponding extensions.
 */
export function extensionsAsCoreFeaturesEnabled(settingsCascade: SettingsCascadeOrError): boolean {
    if (!settingsCascade.final) {
        return false
    }
    if (isErrorLike(settingsCascade.final)) {
        return false
    }

    return settingsCascade.final.experimentalFeatures?.extensionsAsCoreFeatures === true
}

/**
 * Returns undefined if the settings cannot be loaded or if the setting doesn't
 * exist.
 */
function getFromSettings<T>(settingsCascade: SettingsCascadeOrError, setting: string): T | undefined {
    if (!settingsCascade.final) {
        return undefined
    }
    if (isErrorLike(settingsCascade.final)) {
        return undefined
    }

    const value = settingsCascade.final[setting]
    if (value !== undefined) {
        return value as T
    }

    return undefined
}
