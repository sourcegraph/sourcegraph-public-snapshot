import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import React from 'react'
import { isExtensionEnabled } from '../../../shared/src/extensions/extension'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps, SettingsSubject } from '../../../shared/src/settings/settings'
import { isErrorLike } from '../../../shared/src/util/errors'
import { ExtensionCard } from './ExtensionCard'
import { ErrorAlert } from '../components/alerts'
import { applyExtensionsEnablement } from './extensions'
import { ExtensionCategory, EXTENSION_CATEGORIES } from '../../../shared/src/schema/extensionSchema'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'
import { ExtensionListData, ExtensionsEnablement } from './ExtensionRegistry'

interface Props
    extends SettingsCascadeProps,
        PlatformContextProps<'settings' | 'updateSettings' | 'requestGraphQL'>,
        Pick<ExtensionsAreaRouteContext, 'authenticatedUser'> {
    subject: Pick<SettingsSubject, 'id' | 'viewerCanAdminister'>
    location: H.Location
    history: H.History

    data: ExtensionListData | undefined
    selectedCategories: ExtensionCategory[]
    enablementFilter: ExtensionsEnablement
    query: string
    showMoreExtensions: boolean
}

const LOADING = 'loading' as const

/**
 * Displays a list of extensions.
 */
export const ExtensionsList: React.FunctionComponent<Props> = ({
    history,
    subject,
    settingsCascade,
    platformContext,
    data,
    selectedCategories,
    enablementFilter,
    query,
    showMoreExtensions,
}) => {
    if (!data || data === LOADING) {
        return <LoadingSpinner className="icon-inline" />
    }

    if (isErrorLike(data)) {
        return <ErrorAlert error={data} history={history} />
    }

    const { error, categories, extensions } = data

    if (Object.keys(extensions).length === 0) {
        return (
            <>
                {error && <ErrorAlert className="mb-2" error={error} history={history} />}
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

    // Don't display programming language extensions by default
    const renderLanguages =
        (selectedCategories.length === 0 && showMoreExtensions) || selectedCategories.includes('Programming languages')

    const filteredCategoryIDs = EXTENSION_CATEGORIES.filter(category => {
        if (category === 'Programming languages') {
            return renderLanguages
        }

        return selectedCategories.length === 0 || selectedCategories.includes(category)
    })

    // Apply enablement filter
    const filteredCategories = applyExtensionsEnablement(
        categories,
        filteredCategoryIDs,
        enablementFilter,
        settingsCascade.final
    )

    const categorySections = filteredCategoryIDs
        .filter(category => filteredCategories[category].length > 0)
        .map(category => (
            <div key={category} className="mt-1">
                <h3 className="extensions-list__category font-weight-bold">{category}</h3>
                <div className="extensions-list__cards mt-1">
                    {filteredCategories[category].map(extensionId => (
                        <ExtensionCard
                            key={extensionId}
                            subject={subject}
                            node={extensions[extensionId]}
                            settingsCascade={settingsCascade}
                            platformContext={platformContext}
                            enabled={isExtensionEnabled(settingsCascade.final, extensionId)}
                        />
                    ))}
                </div>
            </div>
        ))

    return (
        <>
            {error && <ErrorAlert className="mb-2" error={error} history={history} />}
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
