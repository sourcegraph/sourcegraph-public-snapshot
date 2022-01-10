import classNames from 'classnames'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Form } from 'reactstrap'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { SearchBox } from '@sourcegraph/branded/src/search/input/SearchBox'
import { getFullQuery } from '@sourcegraph/branded/src/search/input/toggles/Toggles'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { AuthenticatedUser, currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'
import { dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { getAvailableSearchContextSpecOrDefault } from '@sourcegraph/shared/src/search'
import {
    fetchAutoDefinedSearchContexts,
    fetchSearchContexts,
    getUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/search/backend'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { globbingEnabledFromSettings } from '@sourcegraph/shared/src/util/globbing'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { BrandLogo } from '@sourcegraph/web/src/components/branding/BrandLogo'
import { TreeEntriesResult, TreeEntriesVariables } from '@sourcegraph/web/src/graphql-operations'
import { SearchBetaIcon } from '@sourcegraph/web/src/search/CtaIcons'

import { SearchResult, SearchVariables } from '../../graphql-operations'
import { LocalRecentSeachProps } from '../contract'
import { WebviewPageProps } from '../platform/context'

import { HomePanels } from './HomePanels'
import styles from './index.module.scss'
import { searchQuery, treeEntriesQuery } from './queries'
import { RepoPage } from './RepoPage'
import { convertGQLSearchToSearchMatches, SearchResults } from './SearchResults'
import { DEFAULT_SEARCH_CONTEXT_SPEC } from './state'

import { useQueryState } from '.'

interface SearchPageProps extends WebviewPageProps {}

export const SearchPage: React.FC<SearchPageProps> = ({ platformContext, theme, sourcegraphVSCodeExtensionAPI }) => {
    const searchActions = useQueryState(({ actions }) => actions)
    const queryState = useQueryState(({ state }) => state.queryState)
    const queryToRun = useQueryState(({ state }) => state.queryToRun)
    const caseSensitive = useQueryState(({ state }) => state.caseSensitive)
    const patternType = useQueryState(({ state }) => state.patternType)
    const selectedSearchContextSpec = useQueryState(({ state }) => state.selectedSearchContextSpec)
    const [fullQuery, setFullQuery] = useState<string | undefined>(undefined)
    const instanceHostname = useMemo(() => sourcegraphVSCodeExtensionAPI.getInstanceHostname(), [
        sourcegraphVSCodeExtensionAPI,
    ])
    const [hasAccessToken, setHasAccessToken] = useState<boolean | undefined>(undefined)
    const [validAccessToken, setValidAccessToken] = useState<boolean>(false)
    const [lastSelectedSearchContext, setLastSelectedSearchContext] = useState<string | undefined>(undefined)
    const [localRecentSearches, setLocalRecentSearches] = useState<LocalRecentSeachProps[] | undefined>(undefined)
    // File Tree
    const [openRepoFileTree, setOpenRepoFileTree] = useState<boolean>(false)
    const [fileVariables, setFileVariables] = useState<TreeEntriesVariables | undefined>(undefined)
    const [entries, setEntries] = useState<Pick<GQL.ITreeEntry, 'name' | 'isDirectory' | 'url' | 'path'>[] | undefined>(
        undefined
    )
    const [authenticatedUser, setAuthenticatedUser] = useState<AuthenticatedUser | null>(null)
    const sourcegraphSettings =
        useObservable(
            useMemo(() => wrapRemoteObservable(sourcegraphVSCodeExtensionAPI.getSettings()), [
                sourcegraphVSCodeExtensionAPI,
            ])
        ) ?? EMPTY_SETTINGS_CASCADE

    const globbing = useMemo(() => globbingEnabledFromSettings(sourcegraphSettings), [sourcegraphSettings])

    const [loading, setLoading] = useState(false)

    const onSubmit = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            setOpenRepoFileTree(false)
            searchActions.submitQuery()
        },
        [searchActions]
    )

    const fetchSuggestions = useCallback(
        (query: string): Observable<SearchMatch[]> =>
            platformContext
                .requestGraphQL<SearchResult, SearchVariables>({
                    request: searchQuery,
                    variables: { query, patternType: null },
                    mightContainPrivateInfo: true,
                })
                .pipe(
                    map(dataOrThrowErrors),
                    map(results => convertGQLSearchToSearchMatches(results)),
                    catchError(() => [])
                ),
        [platformContext]
    )

    const setSelectedSearchContextSpec = (spec: string): void => {
        setLastSelectedSearchContext(spec)
        getAvailableSearchContextSpecOrDefault({
            spec,
            defaultSpec: DEFAULT_SEARCH_CONTEXT_SPEC,
            platformContext,
        })
            .toPromise()
            .then(availableSearchContextSpecOrDefault => {
                searchActions.setSelectedSearchContextSpec(availableSearchContextSpecOrDefault)
            })
            .catch(() => {
                // TODO error handling
            })

        sourcegraphVSCodeExtensionAPI
            .updateLastSelectedSearchContext(spec)
            .then(response => console.log(response))
            .catch(error => console.log(error))
    }

    const getFiles = (variables: TreeEntriesVariables): void => {
        setFileVariables(variables)
        setOpenRepoFileTree(true)
    }

    const onSignUpClick = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            platformContext.telemetryService.log(
                'VSCESearchPageClicked',
                { campaign: 'Sign up link' },
                { campaign: 'Sign up link' }
            )
        },
        [platformContext.telemetryService]
    )

    useEffect(() => {
        // Check for Access Token to display sign up CTA
        if (hasAccessToken === undefined) {
            sourcegraphVSCodeExtensionAPI
                .hasAccessToken()
                .then(hasAccessToken => {
                    setHasAccessToken(hasAccessToken)
                })
                // TODO error handling
                .catch(() => setHasAccessToken(false))
        }
        if (!validAccessToken) {
            ;(async () => {
                const currentUser = await platformContext
                    .requestGraphQL<CurrentAuthStateResult, CurrentAuthStateVariables>({
                        request: currentAuthStateQuery,
                        variables: {},
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                if (currentUser.data) {
                    setAuthenticatedUser(currentUser.data.currentUser)
                    setValidAccessToken(true)
                } else {
                    setValidAccessToken(false)
                }
            })().catch(error => console.error(error))
        }

        // Get Recent Search History from Local Storage
        if (localRecentSearches === undefined) {
            sourcegraphVSCodeExtensionAPI
                .getLocalRecentSearch()
                .then(response => {
                    setLocalRecentSearches(response)
                })
                .catch(() => {
                    // TODO error handling
                })
        }
        if (lastSelectedSearchContext === undefined) {
            setLoading(true)
            sourcegraphVSCodeExtensionAPI
                .getLastSelectedSearchContext()
                .then(spec => {
                    setLastSelectedSearchContext(spec)
                    getAvailableSearchContextSpecOrDefault({
                        spec,
                        defaultSpec: DEFAULT_SEARCH_CONTEXT_SPEC,
                        platformContext,
                    })
                        .toPromise()
                        .then(availableSearchContextSpecOrDefault => {
                            searchActions.setSelectedSearchContextSpec(availableSearchContextSpecOrDefault)
                        })
                        .catch(() => {
                            // TODO error handling
                        })
                })
                // TODO error handling
                .catch(error => console.log(error))
            setLoading(false)
        }

        const subscriptions = new Subscription()

        // create file tree
        if (openRepoFileTree && fileVariables) {
            ;(async () => {
                const files = await platformContext
                    .requestGraphQL<TreeEntriesResult, TreeEntriesVariables>({
                        request: treeEntriesQuery,
                        variables: fileVariables,
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                if (files.data?.repository?.commit?.tree) {
                    setEntries(files.data.repository.commit.tree.entries)
                }
            })().catch(error => console.error(error))
        }

        if (queryToRun.query) {
            setLoading(true)

            const currentFullQuery = getFullQuery(
                queryToRun.query,
                selectedSearchContextSpec || '',
                caseSensitive,
                patternType
            )
            setFullQuery(currentFullQuery)

            let queryString = `${queryToRun.query}${caseSensitive ? ' case:yes' : ''}`

            if (selectedSearchContextSpec) {
                queryString = appendContextFilter(queryString, selectedSearchContextSpec)
            }

            if (fullQuery && localRecentSearches !== undefined && localRecentSearches.length < 12) {
                // query to add to search history
                const newSearchHistory = {
                    lastQuery: queryToRun.query,
                    lastSelectedSearchContextSpec: selectedSearchContextSpec || '',
                    lastCaseSensitive: caseSensitive,
                    lastPatternType: patternType,
                    lastFullQuery: fullQuery,
                }
                if (localRecentSearches[localRecentSearches.length - 1]?.lastFullQuery !== fullQuery) {
                    let currentLocalSearchHistory = localRecentSearches
                    // Local Search History is limited to 10
                    if (localRecentSearches.length > 9) {
                        currentLocalSearchHistory = localRecentSearches.slice(-9)
                    }
                    const newRecentSearches = [...currentLocalSearchHistory, newSearchHistory]
                    setLocalRecentSearches(newRecentSearches)
                    sourcegraphVSCodeExtensionAPI
                        .setLocalRecentSearch(newRecentSearches)
                        .then(response => {
                            console.log('Added to search history', response)
                        }) // TODO error handling
                        .catch(error => console.log(error))
                }
            }

            const subscription = platformContext
                .requestGraphQL<SearchResult, SearchVariables>({
                    request: searchQuery,
                    variables: { query: queryString, patternType },
                    mightContainPrivateInfo: true,
                })
                .pipe(map(dataOrThrowErrors)) // TODO error handling
                .subscribe(searchResults => {
                    searchActions.updateResults(searchResults)
                    setLoading(false)
                })

            subscriptions.add(subscription)
        }

        return () => subscriptions.unsubscribe()
    }, [
        sourcegraphVSCodeExtensionAPI,
        queryToRun,
        patternType,
        caseSensitive,
        selectedSearchContextSpec,
        searchActions,
        platformContext,
        hasAccessToken,
        lastSelectedSearchContext,
        localRecentSearches,
        fullQuery,
        fileVariables,
        instanceHostname,
        validAccessToken,
        openRepoFileTree,
    ])

    return (
        <div>
            {!queryToRun.query ? (
                <div className={classNames('d-flex flex-column align-items-center px-3', styles.searchPage)}>
                    <BrandLogo
                        className={styles.logo}
                        isLightTheme={theme === 'theme-light'}
                        variant="logo"
                        assetsRoot="https://sourcegraph.com/.assets"
                    />
                    <div className="text-muted text-center font-italic mt-3">
                        Search your code and 2M+ open source repositories
                    </div>
                    <div className={classNames(styles.searchContainer, styles.searchContainerWithContentBelow)}>
                        {!loading && (
                            <Form className="d-flex my-2" onSubmit={onSubmit}>
                                {/* TODO temporary settings provider w/ mock in memory storage */}
                                <SearchBox
                                    isSourcegraphDotCom={true}
                                    // Platform context props
                                    platformContext={platformContext}
                                    telemetryService={platformContext.telemetryService}
                                    // Search context props
                                    searchContextsEnabled={true}
                                    showSearchContext={true}
                                    showSearchContextManagement={true}
                                    hasUserAddedExternalServices={false}
                                    hasUserAddedRepositories={true} // Used for search context CTA, which we won't show here.
                                    defaultSearchContextSpec={DEFAULT_SEARCH_CONTEXT_SPEC}
                                    // TODO store search context in vs code settings?
                                    setSelectedSearchContextSpec={setSelectedSearchContextSpec}
                                    selectedSearchContextSpec={selectedSearchContextSpec}
                                    fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                                    fetchSearchContexts={fetchSearchContexts}
                                    getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                                    // Case sensitivity props
                                    caseSensitive={caseSensitive}
                                    setCaseSensitivity={searchActions.setCaseSensitivity}
                                    // Pattern type props
                                    patternType={patternType}
                                    setPatternType={searchActions.setPatternType}
                                    // Misc.
                                    isLightTheme={theme === 'theme-light'}
                                    authenticatedUser={null} // Used for search context CTA, which we won't show here.
                                    queryState={queryState}
                                    onChange={searchActions.setQuery}
                                    onSubmit={onSubmit}
                                    autoFocus={true}
                                    fetchSuggestions={fetchSuggestions}
                                    settingsCascade={sourcegraphSettings}
                                    globbing={globbing}
                                    // TODO(tj): instead of cssvar, can pipe in font settings from extension
                                    // to be able to pass it to Monaco!
                                    className={classNames(
                                        styles.withEditorFont,
                                        'flex-grow-1 flex-shrink-past-contents'
                                    )}
                                />
                            </Form>
                        )}
                    </div>
                    <div className="flex-grow-1">
                        <HomePanels
                            telemetryService={platformContext.telemetryService}
                            isLightTheme={theme === 'theme-light'}
                            setQuery={searchActions.setQuery}
                        />
                    </div>
                </div>
            ) : (
                <>
                    <Form className="d-flex my-2" onSubmit={onSubmit}>
                        {/* TODO temporary settings provider w/ mock in memory storage */}
                        <SearchBox
                            isSourcegraphDotCom={true}
                            // Platform context props
                            platformContext={platformContext}
                            telemetryService={platformContext.telemetryService}
                            // Search context props
                            searchContextsEnabled={true}
                            showSearchContext={true}
                            showSearchContextManagement={true}
                            hasUserAddedExternalServices={false}
                            hasUserAddedRepositories={true} // Used for search context CTA, which we won't show here.
                            defaultSearchContextSpec={DEFAULT_SEARCH_CONTEXT_SPEC}
                            // TODO store search context in vs code settings?
                            setSelectedSearchContextSpec={setSelectedSearchContextSpec}
                            selectedSearchContextSpec={selectedSearchContextSpec}
                            fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                            fetchSearchContexts={fetchSearchContexts}
                            getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                            // Case sensitivity props
                            caseSensitive={caseSensitive}
                            setCaseSensitivity={searchActions.setCaseSensitivity}
                            // Pattern type props
                            patternType={patternType}
                            setPatternType={searchActions.setPatternType}
                            // Misc.
                            isLightTheme={theme === 'theme-light'}
                            authenticatedUser={null} // Used for search context CTA, which we won't show here.
                            queryState={queryState}
                            onChange={searchActions.setQuery}
                            onSubmit={onSubmit}
                            autoFocus={true}
                            fetchSuggestions={fetchSuggestions}
                            settingsCascade={sourcegraphSettings}
                            globbing={globbing}
                            // TODO(tj): instead of cssvar, can pipe in font settings from extension
                            // to be able to pass it to Monaco!
                            className={classNames(styles.withEditorFont, 'flex-grow-1 flex-shrink-past-contents')}
                        />
                    </Form>
                    {loading ? (
                        <p>Loading...</p>
                    ) : (
                        // Display Sign up banner if no access token is detected (assuming they do not have a Sourcegraph account)
                        <div className={classNames(styles.streamingSearchResultsContainer)}>
                            {!hasAccessToken && (
                                <div className="card my-2 mr-3 d-flex p-3 flex-md-row flex-column align-items-center">
                                    <div className="mr-md-3">
                                        <SearchBetaIcon />
                                    </div>
                                    <div
                                        className={classNames(
                                            'flex-1 my-md-0 my-2',
                                            styles.streamingSearchResultsCtaContainer
                                        )}
                                    >
                                        <div className={classNames('mb-1', styles.streamingSearchResultsCtaTitle)}>
                                            <strong>
                                                Sign up to add your public and private repositories and access other
                                                features
                                            </strong>
                                        </div>
                                        <div
                                            className={classNames(
                                                'text-muted',
                                                styles.streamingSearchResultsCtaDescription
                                            )}
                                        >
                                            Do all the things editors can’t: search multiple repos & commit history,
                                            monitor, save searches and more.
                                        </div>
                                    </div>
                                    <a
                                        className={classNames('btn', styles.streamingSearchResultsBtn)}
                                        href="https://sourcegraph.com/sign-up?src=SearchCTA"
                                        onClick={onSignUpClick}
                                    >
                                        <span className={styles.streamingSearchResultsText}>Create a free account</span>
                                    </a>
                                </div>
                            )}
                            {/* TODO: This is a temporary repo file viewer */}
                            {openRepoFileTree && fileVariables && entries !== undefined && (
                                <section className={classNames('test-tree-entries mb-3')}>
                                    <h2>Files and directories</h2>
                                    <RepoPage
                                        platformContext={platformContext}
                                        theme={theme}
                                        getFiles={getFiles}
                                        entries={entries}
                                        instanceHostname={instanceHostname}
                                        sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                                        selectedRepoName={fileVariables.repoName}
                                    />
                                </section>
                            )}
                            {fullQuery && !openRepoFileTree && (
                                <SearchResults
                                    platformContext={platformContext}
                                    theme={theme}
                                    sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                                    settings={sourcegraphSettings}
                                    instanceHostname={instanceHostname}
                                    fullQuery={fullQuery}
                                    getFiles={getFiles}
                                    authenticatedUser={authenticatedUser}
                                />
                            )}
                        </div>
                    )}
                </>
            )}
        </div>
    )
}
