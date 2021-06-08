import * as H from 'history'
import React, { useMemo } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { isExtensionEnabled } from '@sourcegraph/shared/src/extensions/extension'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { ExtensionCategory, EXTENSION_CATEGORIES } from '@sourcegraph/shared/src/schema/extensionSchema'
import { SettingsCascadeProps, SettingsSubject } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { createRecord } from '@sourcegraph/shared/src/util/createRecord'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../components/alerts'

import { ExtensionCard } from './ExtensionCard'
import { ExtensionCategoryOrAll, ExtensionListData, ExtensionsEnablement } from './ExtensionRegistry'
import { applyEnablementFilter, applyWIPFilter } from './extensions'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'

interface Props
    extends SettingsCascadeProps,
        PlatformContextProps<'settings' | 'updateSettings' | 'requestGraphQL'>,
        Pick<ExtensionsAreaRouteContext, 'authenticatedUser'>,
        ThemeProps {
    subject: Pick<SettingsSubject, 'id' | 'viewerCanAdminister'>
    location: H.Location

    data: ExtensionListData | undefined
    selectedCategory: ExtensionCategoryOrAll
    enablementFilter: ExtensionsEnablement
    showExperimentalExtensions: boolean
    query: string
    onShowFullCategoryClicked: (category: ExtensionCategory) => void
}

const LOADING = 'loading' as const

/** Categories, but with 'Programming Languages' at the end */
const ORDERED_EXTENSION_CATEGORIES: ExtensionCategory[] = [
    ...EXTENSION_CATEGORIES.filter(category => category !== 'Programming languages'),
    'Programming languages',
]

/**
 * Displays a list of extensions.
 */
export const ExtensionsList: React.FunctionComponent<Props> = ({
    subject,
    settingsCascade,
    platformContext,
    data,
    selectedCategory,
    enablementFilter,
    showExperimentalExtensions,
    query,
    authenticatedUser,
    onShowFullCategoryClicked,
    ...props
}) => {
    // Settings subjects for extension toggle
    const viewerSubject = useMemo(
        () => settingsCascade.subjects?.find(settingsSubject => settingsSubject.subject.id === subject.id),
        [settingsCascade, subject.id]
    )
    const siteSubject = useMemo(() => {
        if (!settingsCascade.subjects) {
            return undefined
        }
        for (const subject of settingsCascade.subjects) {
            if (subject.subject.__typename === 'Site') {
                // Even if the user has permission to administer Site settings, changes cannot be made
                // through the API if global settings are configured through the GLOBAL_SETTINGS_FILE envvar.
                if (subject.subject.allowSiteSettingsEdits) {
                    return subject
                }
                break
            }
        }
        return undefined
    }, [settingsCascade])

    if (!data || data === LOADING) {
        return <LoadingSpinner className="icon-inline mt-2" />
    }

    if (isErrorLike(data)) {
        return <ErrorAlert error={data} />
    }

    const { error, extensions, extensionIDsByCategory } = data

    if (Object.keys(extensions).length === 0) {
        return (
            <>
                {error && <ErrorAlert className="mb-2" error={error} />}
                {query ? (
                    <div className="text-muted">
                        No extensions match <strong>{query}</strong>.
                    </div>
                ) : (
                    <span className="text-muted">No extensions found</span>
                )}
            </>
        )
    }

    let categorySections: JSX.Element[]

    if (selectedCategory === 'All') {
        const filteredExtensionIDsByCategory = createRecord(ORDERED_EXTENSION_CATEGORIES, category => {
            const enablementFilteredExtensionIDs = applyEnablementFilter(
                extensionIDsByCategory[category].primaryExtensionIDs,
                enablementFilter,
                settingsCascade.final
            )
            return applyWIPFilter(enablementFilteredExtensionIDs, showExperimentalExtensions, extensions)
        })

        categorySections = ORDERED_EXTENSION_CATEGORIES.filter(
            category =>
                filteredExtensionIDsByCategory[category].length > 0 &&
                // Only show Programming Languages when it is the selected category
                category !== 'Programming languages'
        ).map(category => {
            const extensionIDsForCategory = filteredExtensionIDsByCategory[category]

            return (
                <div key={category} className="mt-1">
                    <h3
                        className="extensions-list__category mb-3 font-weight-normal"
                        data-test-extension-category-header={category}
                    >
                        {category}
                    </h3>
                    <div className="extensions-list__cards mt-1">
                        {extensionIDsForCategory.slice(0, 6).map(extensionId => (
                            <ExtensionCard
                                key={extensionId}
                                subject={subject}
                                viewerSubject={viewerSubject?.subject}
                                siteSubject={siteSubject?.subject}
                                node={extensions[extensionId]}
                                settingsCascade={settingsCascade}
                                platformContext={platformContext}
                                enabled={isExtensionEnabled(settingsCascade.final, extensionId)}
                                enabledForAllUsers={
                                    siteSubject ? isExtensionEnabled(siteSubject.settings, extensionId) : false
                                }
                                isLightTheme={props.isLightTheme}
                                settingsURL={authenticatedUser?.settingsURL}
                                authenticatedUser={authenticatedUser}
                            />
                        ))}
                    </div>
                    {extensionIDsForCategory.length > 6 && (
                        <div className="d-flex justify-content-center mt-4">
                            <button
                                type="button"
                                className="btn btn-outline-secondary"
                                onClick={() => onShowFullCategoryClicked(category)}
                            >
                                Show full category
                            </button>
                        </div>
                    )}
                </div>
            )
        })
    } else {
        // When a category is selected, display all extensions that include this category in their manifest,
        // not just extensions for which this is the primary category.
        const { allExtensionIDs } = extensionIDsByCategory[selectedCategory]

        const enablementFilteredExtensionIDs = applyEnablementFilter(
            allExtensionIDs,
            enablementFilter,
            settingsCascade.final
        )
        const extensionIDs = applyWIPFilter(enablementFilteredExtensionIDs, showExperimentalExtensions, extensions)

        categorySections = [
            <div key={selectedCategory} className="mt-1">
                <h3
                    className="extensions-list__category mb-3 font-weight-normal"
                    data-test-extension-category-header={selectedCategory}
                >
                    {selectedCategory}
                </h3>
                <div className="extensions-list__cards mt-1">
                    {extensionIDs.map(extensionId => (
                        <ExtensionCard
                            key={extensionId}
                            subject={subject}
                            viewerSubject={viewerSubject?.subject}
                            siteSubject={siteSubject?.subject}
                            node={extensions[extensionId]}
                            settingsCascade={settingsCascade}
                            platformContext={platformContext}
                            enabled={isExtensionEnabled(settingsCascade.final, extensionId)}
                            enabledForAllUsers={
                                siteSubject ? isExtensionEnabled(siteSubject.settings, extensionId) : false
                            }
                            isLightTheme={props.isLightTheme}
                            settingsURL={authenticatedUser?.settingsURL}
                            authenticatedUser={authenticatedUser}
                        />
                    ))}
                </div>
            </div>,
        ]
    }

    return (
        <>
            {error && <ErrorAlert className="mb-2" error={error} />}
            {categorySections.length > 0 ? (
                categorySections
            ) : (
                <div className="text-muted">
                    No extensions match <strong>{query}</strong> in the selected categories.
                </div>
            )}
        </>
    )
}
