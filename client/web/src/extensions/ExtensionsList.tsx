import React, { useMemo } from 'react'

import classNames from 'classnames'
import * as H from 'history'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { isExtensionEnabled } from '@sourcegraph/shared/src/extensions/extension'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { ExtensionCategory, EXTENSION_CATEGORIES } from '@sourcegraph/shared/src/schema/extensionSchema'
import { mergeSettings, SettingsCascadeProps, SettingsSubject } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, LoadingSpinner, Typography } from '@sourcegraph/wildcard'

import { ExtensionCard } from './ExtensionCard'
import { ExtensionCategoryOrAll, ExtensionListData, ExtensionsEnablement } from './ExtensionRegistry'
import { applyEnablementFilter, applyWIPFilter } from './extensions'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'
import { createRecord } from './utils/createRecord'

import styles from './ExtensionsList.module.scss'

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
export const ExtensionsList: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
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
                    // Merge default settings (to include e.g. programming language extension settings that may not
                    // have been modified by site admins).
                    const defaultSubject = settingsCascade.subjects.find(
                        subject => subject.subject.__typename === 'DefaultSettings'
                    )
                    const defaultSettings =
                        !!defaultSubject?.settings && !isErrorLike(defaultSubject.settings)
                            ? defaultSubject.settings
                            : undefined

                    if (!!subject.settings && !isErrorLike(subject.settings) && defaultSettings) {
                        // Site settings have higher precedence than default settings, so put them
                        // after default settings in the array.
                        const mergedSettings = mergeSettings([defaultSettings, subject.settings])
                        return { ...subject, settings: mergedSettings }
                    }

                    return subject
                }
                break
            }
        }
        return undefined
    }, [settingsCascade])

    // We want to filter extensions based on the settings when the filter was last changed.
    // If we used the latest settings, which are updated after extensions are enabled/disabled,
    // extensions would confusingly disappear when a filter is used (example: user wants to see disabled extensions.
    // user enables an extension, then the extension disappears since it is now enabled in settings).
    // eslint-disable-next-line react-hooks/exhaustive-deps
    const settingsFromLastFilterChange = useMemo(() => settingsCascade.final, [enablementFilter])

    if (!data || data === LOADING) {
        return <LoadingSpinner className="mt-2" />
    }

    if (isErrorLike(data)) {
        return <ErrorAlert error={data} />
    }

    const { error, extensions, extensionIDsByCategory, featuredExtensions } = data

    const featuredExtensionsSection = featuredExtensions && featuredExtensions.length > 0 && (
        <div key="Featured" className={styles.featuredSection}>
            <Typography.H3
                className={classNames('mb-3 font-weight-normal', styles.category)}
                data-test-extension-category-header="Featured"
            >
                Featured
            </Typography.H3>
            <div className={classNames('mt-1', styles.cards, styles.cardsFeatured)}>
                {featuredExtensions.map(featuredExtension => (
                    <ExtensionCard
                        key={featuredExtension.id}
                        subject={subject}
                        viewerSubject={viewerSubject?.subject}
                        siteSubject={siteSubject?.subject}
                        node={featuredExtension}
                        settingsCascade={settingsCascade}
                        platformContext={platformContext}
                        enabled={isExtensionEnabled(settingsCascade.final, featuredExtension.id)}
                        enabledForAllUsers={
                            siteSubject ? isExtensionEnabled(siteSubject.settings, featuredExtension.id) : false
                        }
                        isLightTheme={props.isLightTheme}
                        settingsURL={authenticatedUser?.settingsURL}
                        authenticatedUser={authenticatedUser}
                        featured={true}
                    />
                ))}
            </div>
        </div>
    )

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
                settingsFromLastFilterChange
            )
            return applyWIPFilter(enablementFilteredExtensionIDs, showExperimentalExtensions, extensions)
        })

        categorySections = ORDERED_EXTENSION_CATEGORIES.filter(category => {
            // Hide category if there are no results
            if (filteredExtensionIDsByCategory[category].length === 0) {
                return false
            }
            // Only show Programming Languages when it is the selected category
            // AND there is no search query. We shouldn't hide language extensions
            // if users are looking for them.
            return category !== 'Programming languages' || !!query.trim()
        }).map(category => {
            const extensionIDsForCategory = filteredExtensionIDsByCategory[category]

            return (
                <div key={category} className="mt-1">
                    <Typography.H3
                        className={classNames('mb-3 font-weight-normal', styles.category)}
                        data-test-extension-category-header={category}
                    >
                        {category}
                    </Typography.H3>
                    <div className={classNames('mt-1', styles.cards)}>
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
                            <Button
                                onClick={() => onShowFullCategoryClicked(category)}
                                outline={true}
                                variant="secondary"
                            >
                                Show full category
                            </Button>
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
            settingsFromLastFilterChange
        )
        const extensionIDs = applyWIPFilter(enablementFilteredExtensionIDs, showExperimentalExtensions, extensions)

        categorySections = [
            <div key={selectedCategory} className="mt-1">
                <Typography.H3
                    className={classNames('mb-3 font-weight-normal', styles.category)}
                    data-test-extension-category-header={selectedCategory}
                >
                    {selectedCategory}
                </Typography.H3>
                <div className={classNames('mt-1', styles.cards)}>
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
            {enablementFilter === 'all' && featuredExtensionsSection}
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
