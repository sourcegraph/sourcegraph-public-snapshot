import classNames from 'classnames'
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
                {validAccessToken && savedSearch && (
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
                {!validAccessToken && (
                    <div className={styles.sidebarSection}>
                        <h5 className="flex-grow-1 btn-outline-secondary my-2">Search Your Private Code</h5>
                        <div className={classNames('p-1', styles.sidebarSectionCta)}>
                            <div className={classNames('p-1', styles.sidebarSectionListItem)}>
                                <p className={classNames('mt-1 mb-2 text')}>
                                    Create an account to enhance search across your private repositories: search
                                    multiple repos & commit history, monitor, save searches, and more.
                                </p>
                            </div>
                            <div className={classNames('p-1 m-0', styles.sidebarSectionListItem)}>
                                <button
                                    type="submit"
                                    className={classNames(
                                        'btn btn-sm btn-primary btn-link w-100 border-0 font-weight-normal'
                                    )}
                                >
                                    <span className="py-1">Create an account</span>
                                </button>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        )
    }
    return <LoadingSpinner />
}
