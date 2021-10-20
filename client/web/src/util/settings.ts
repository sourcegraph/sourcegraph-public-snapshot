import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { AuthenticatedUser } from '../auth'
import { LayoutProps } from '../Layout'
import { SettingsExperimentalFeatures } from '../schema/settings.schema'
import { parseSearchURLPatternType } from '../search'

/** A fallback settings subject that can be constructed synchronously at initialization time. */
export const SITE_SUBJECT_NO_ADMIN: Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'> = {
    id: window.context.siteGQLID,
    viewerCanAdminister: false,
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
    return SITE_SUBJECT_NO_ADMIN
}

export function defaultPatternTypeFromSettings(settingsCascade: SettingsCascadeOrError): SearchPatternType | undefined {
    // When the web app mounts, if the current page does not have a patternType URL
    // parameter, set the search pattern type to the defaultPatternType from settings
    // (if it is set), otherwise default to literal.
    //
    // For search result URLs that have no patternType= query parameter,
    // the `SearchResults` component will append &patternType=regexp
    // to the URL to ensure legacy search links continue to work.
    if (!parseSearchURLPatternType(window.location.search)) {
        const defaultPatternType =
            settingsCascade.final &&
            !isErrorLike(settingsCascade.final) &&
            (settingsCascade.final['search.defaultPatternType'] as SearchPatternType.literal)
        return defaultPatternType || SearchPatternType.literal
    }
    return
}

export function defaultCaseSensitiveFromSettings(settingsCascade: SettingsCascadeOrError): boolean {
    // Analogous to defaultPatternTypeFromSettings, but for case sensitivity.
    if (!parseSearchURLPatternType(window.location.search)) {
        const defaultCaseSensitive =
            settingsCascade.final &&
            !isErrorLike(settingsCascade.final) &&
            (settingsCascade.final['search.defaultCaseSensitive'] as boolean)
        return defaultCaseSensitive || false
    }
    return false
}

export function experimentalFeaturesFromSettings(
    settingsCascade: SettingsCascadeOrError
): {
    showOnboardingTour: boolean
    showEnterpriseHomePanels: boolean
    showMultilineSearchConsole: boolean
    showSearchNotebook: boolean
    showSearchContext: boolean
    showSearchContextManagement: boolean
    enableCodeMonitoring: boolean
    enableAPIDocs: boolean
} {
    const experimentalFeatures: SettingsExperimentalFeatures =
        (settingsCascade.final && !isErrorLike(settingsCascade.final) && settingsCascade.final.experimentalFeatures) ||
        {}

    const {
        showOnboardingTour = true, // Default to true if not set
        showEnterpriseHomePanels = true, // Default to true if not set
        showSearchContext = true, // Default to true if not set
        showSearchContextManagement = true, // Default to true if not set
        showMultilineSearchConsole = false,
        showSearchNotebook = false,
        codeMonitoring = true, // Default to true if not set
        // eslint-disable-next-line unicorn/prevent-abbreviations
        apiDocs = true, // Default to true if not set
    } = experimentalFeatures

    return {
        showOnboardingTour,
        showSearchContext,
        showSearchContextManagement,
        showEnterpriseHomePanels,
        showMultilineSearchConsole,
        showSearchNotebook,
        enableCodeMonitoring: codeMonitoring,
        enableAPIDocs: apiDocs,
    }
}
