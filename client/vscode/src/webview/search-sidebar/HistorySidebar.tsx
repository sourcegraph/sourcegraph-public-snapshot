import React, { useEffect, useState } from 'react'
import { UseStore } from 'zustand'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { ISavedSearch } from '@sourcegraph/shared/src/graphql/schema'
import { SearchQueryState } from '@sourcegraph/shared/src/search/searchQueryState'

import { SearchPatternType } from '../../graphql-operations'
import { LocalRecentSeachProps } from '../contract'
import { WebviewPageProps } from '../platform/context'
import { savedSearchQuery } from '../search-panel/queries'

import styles from './HistorySidebar.module.scss'
import { RecentFile } from './RecentFile'
import { RecentRepo } from './RecentRepo'
import { RecentSearch } from './RecentSearch'
import { SaveSearches } from './SaveSearches'
import { SearchTypes } from './SearchType'

interface HistorySidebarProps extends WebviewPageProps {
    authenticatedUser: AuthenticatedUser | null
    patternType: SearchPatternType
    caseSensitive: boolean
    useQueryState: UseStore<SearchQueryState>
    localRecentSearches: LocalRecentSeachProps[]
    localFileHistory: string[]
    validAccessToken: boolean
}

export const HistorySidebar: React.FC<HistorySidebarProps> = ({
    sourcegraphVSCodeExtensionAPI,
    platformContext,
    theme,
    authenticatedUser,
    patternType,
    caseSensitive,
    useQueryState,
    localRecentSearches,
    localFileHistory,
    validAccessToken,
}) => {
    const [savedSearch, setSavedSearch] = useState<ISavedSearch[] | null | undefined>(undefined)

    useEffect(() => {
        if (savedSearch === undefined && authenticatedUser) {
            ;(async () => {
                const savedSearches = await platformContext
                    .requestGraphQL<{ savedSearches: ISavedSearch[] }>({
                        request: savedSearchQuery,
                        variables: {},
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                if (savedSearches.data) {
                    setSavedSearch(savedSearches.data.savedSearches)
                } else {
                    setSavedSearch(null)
                }
            })().catch(() => setSavedSearch(null))
        }
    }, [
        sourcegraphVSCodeExtensionAPI,
        localRecentSearches,
        authenticatedUser,
        platformContext,
        savedSearch,
        localFileHistory,
    ])

    if (localRecentSearches && authenticatedUser !== undefined && localFileHistory !== undefined) {
        return (
            <div className={styles.sidebarContainer}>
                {useQueryState && (
                    <SearchTypes
                        forceButton={true}
                        useQueryState={useQueryState}
                        patternType={patternType}
                        caseSensitive={caseSensitive}
                        sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                    />
                )}
                {validAccessToken && savedSearch && savedSearch.length > 0 && (
                    <SaveSearches
                        savedSearches={savedSearch}
                        telemetryService={platformContext.telemetryService}
                        platformContext={platformContext}
                        sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                        theme={theme}
                    />
                )}
                <RecentSearch
                    localRecentSearches={localRecentSearches}
                    telemetryService={platformContext.telemetryService}
                    authenticatedUser={authenticatedUser}
                    platformContext={platformContext}
                    sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                    theme={theme}
                />
                <RecentRepo
                    localRecentSearches={localRecentSearches}
                    telemetryService={platformContext.telemetryService}
                    authenticatedUser={authenticatedUser}
                    platformContext={platformContext}
                    sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                    theme={theme}
                />
                <RecentFile
                    localFileHistory={localFileHistory}
                    telemetryService={platformContext.telemetryService}
                    authenticatedUser={authenticatedUser}
                    platformContext={platformContext}
                    sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                    theme={theme}
                />
            </div>
        )
    }
    return <LoadingSpinner />
}
