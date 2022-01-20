import classNames from 'classnames'
import React, { useEffect, useMemo, useState } from 'react'
import create, { UseStore } from 'zustand'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { ISavedSearch } from '@sourcegraph/shared/src/graphql/schema'
import { SearchQueryState, updateQuery } from '@sourcegraph/shared/src/search/searchQueryState'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { CurrentAuthStateResult, CurrentAuthStateVariables, SearchPatternType } from '../../graphql-operations'
import { LocalRecentSeachProps } from '../contract'
import { WebviewPageProps } from '../platform/context'
import { currentAuthStateQuery, savedSearchQuery } from '../search-panel/queries'

import styles from './HistorySidebar.module.scss'
import { RecentFile } from './RecentFile'
import { RecentRepo } from './RecentRepo'
import { RecentSearch } from './RecentSearch'
import { SaveSearches } from './SaveSearches'
import { SearchTypes } from './SearchType'

interface HistorySidebarProps extends WebviewPageProps {}

export const HistorySidebar: React.FC<HistorySidebarProps> = ({
    sourcegraphVSCodeExtensionAPI,
    platformContext,
    theme,
}) => {
    const [localRecentSearches, setLocalRecentSearches] = useState<LocalRecentSeachProps[] | undefined>(undefined)
    const [localFileHistory, setLocalFileHistory] = useState<string[] | undefined>(undefined)
    const [authenticatedUser, setAuthenticatedUser] = useState<AuthenticatedUser | null | undefined>(undefined)
    const [savedSearch, setSavedSearch] = useState<ISavedSearch[] | null | undefined>(undefined)
    // Search Query
    const [patternType, setPatternType] = useState<SearchPatternType>(SearchPatternType.literal)
    const [caseSensitive, setCaseSensitive] = useState<boolean>(false)

    const useQueryState: UseStore<SearchQueryState> = useMemo(() => {
        const useStore = create<SearchQueryState>((set, get) => ({
            queryState: { query: '' },
            setQueryState: queryStateUpdate => {
                const queryState =
                    typeof queryStateUpdate === 'function' ? queryStateUpdate(get().queryState) : queryStateUpdate
                set({ queryState })
                // TODO error handling

                sourcegraphVSCodeExtensionAPI.setActiveWebviewQueryState(queryState).then(
                    () => {},
                    () => {}
                )
            },
            submitSearch: (_parameters, updates = []) => {
                const updatedQuery = updateQuery(get().queryState.query, updates)
                // TODO error handling
                sourcegraphVSCodeExtensionAPI
                    .submitActiveWebviewSearch({
                        query: updatedQuery,
                    })
                    .then(
                        () => {},
                        () => {}
                    )
            },
        }))
        return useStore
    }, [sourcegraphVSCodeExtensionAPI])

    const activeQueryState = useObservable(
        useMemo(() => wrapRemoteObservable(sourcegraphVSCodeExtensionAPI.observeActiveWebviewQueryState()), [
            sourcegraphVSCodeExtensionAPI,
        ])
    )

    useEffect(() => {
        // On changes that originate from user input in the search webview panel itself,
        // we don't want to trigger another query state update, which would lead to an infinite loop.
        // That's why we set the state directly, instead of using the `setQueryState` method which
        // updates query state in the search webview panel.
        if (activeQueryState) {
            useQueryState.setState({ queryState: activeQueryState.queryState })
            setPatternType(activeQueryState.patternType)
            setCaseSensitive(activeQueryState.caseSensitive)
        }
    }, [activeQueryState, sourcegraphVSCodeExtensionAPI, useQueryState])

    useEffect(() => {
        // Get initial settings
        if (localRecentSearches === undefined) {
            // Get Local Search History
            sourcegraphVSCodeExtensionAPI
                .getLocalRecentSearch()
                .then(response => {
                    setLocalRecentSearches(response)
                })
                .catch(() => {
                    // TODO error handling
                })
            sourcegraphVSCodeExtensionAPI
                .getLocalStorageItem('sg-files-history')
                .then(response => {
                    setLocalFileHistory(response)
                })
                .catch(() => {
                    // TODO error handling
                })
        }
        if (authenticatedUser === undefined) {
            ;(async () => {
                const currentUser = await platformContext
                    .requestGraphQL<CurrentAuthStateResult, CurrentAuthStateVariables>({
                        request: currentAuthStateQuery,
                        variables: {},
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                // If user is detected, set valid access token to true
                if (currentUser.data) {
                    setAuthenticatedUser(currentUser.data.currentUser)
                } else {
                    setAuthenticatedUser(null)
                }
            })().catch(() => setAuthenticatedUser(null))
        }
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
    }, [sourcegraphVSCodeExtensionAPI, localRecentSearches, authenticatedUser, platformContext, savedSearch])

    if (localRecentSearches && authenticatedUser !== undefined) {
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
                {authenticatedUser && savedSearch && (
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
                {localFileHistory && localFileHistory.length > 0 && (
                    <RecentFile
                        localFileHistory={localFileHistory}
                        telemetryService={platformContext.telemetryService}
                        authenticatedUser={authenticatedUser}
                        platformContext={platformContext}
                        sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                        theme={theme}
                    />
                )}
                {!authenticatedUser && (
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
