import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
import { Observable } from 'rxjs'
import { useDeepCompareEffectNoCheck } from 'use-deep-compare-effect'

import {
    SearchPatternType,
    fetchAutoDefinedSearchContexts,
    getUserSearchContextNamespaces,
    QueryState,
} from '@sourcegraph/search'
import {
    IEditor,
    SearchBox,
    StreamingProgress,
    StreamingSearchResultsList,
    FetchFileParameters,
} from '@sourcegraph/search-ui'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import { collectMetrics } from '@sourcegraph/shared/src/search/query/metrics'
import {
    appendContextFilter,
    sanitizeQueryForTelemetry,
    updateFilters,
} from '@sourcegraph/shared/src/search/query/transformer'
import { LATEST_VERSION, RepositoryMatch, SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { globbingEnabledFromSettings } from '@sourcegraph/shared/src/util/globbing'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { SearchResultsState } from '../../state'
import { WebviewPageProps } from '../platform/context'

import { fetchSearchContexts } from './alias/fetchSearchContext'
import { setFocusSearchBox } from './api'
import { SavedSearchCreateForm } from './components/SavedSearchForm'
import { SearchResultsInfoBar } from './components/SearchResultsInfoBar'
import { MatchHandlersContext, useMatchHandlers } from './MatchHandlersContext'
import { RepoView } from './RepoView'

import styles from './index.module.scss'

export interface SearchResultsViewProps extends WebviewPageProps {
    context: SearchResultsState['context']
}

export const SearchResultsView: React.FunctionComponent<React.PropsWithChildren<SearchResultsViewProps>> = ({
    extensionCoreAPI,
    authenticatedUser,
    platformContext,
    settingsCascade,
    theme,
    context,
    instanceURL,
}) => {
    const [userQueryState, setUserQueryState] = useState<QueryState>(context.submittedSearchQueryState.queryState)
    const [repoToShow, setRepoToShow] = useState<Pick<
        RepositoryMatch,
        'repository' | 'branches' | 'description'
    > | null>(null)

    const isSourcegraphDotCom = useMemo(() => {
        const hostname = new URL(instanceURL).hostname
        return hostname === 'sourcegraph.com' || hostname === 'www.sourcegraph.com'
    }, [instanceURL])

    // Editor focus.
    const editorReference = useRef<IEditor>()
    const setEditor = useCallback((editor: IEditor) => {
        editorReference.current = editor
        setTimeout(() => editor.focus(), 0)
    }, [])

    // TODO explain
    useEffect(() => {
        setFocusSearchBox(() => editorReference.current?.focus())

        return () => {
            setFocusSearchBox(null)
        }
    }, [])

    const onChange = useCallback(
        (newState: QueryState) => {
            setUserQueryState(newState)

            extensionCoreAPI
                .setSidebarQueryState({
                    queryState: newState,
                    searchCaseSensitivity: context.submittedSearchQueryState?.searchCaseSensitivity,
                    searchPatternType: context.submittedSearchQueryState?.searchPatternType,
                })
                .catch(error => {
                    // TODO surface error to users
                    console.error('Error updating sidebar query state from panel', error)
                })
        },
        [
            extensionCoreAPI,
            context.submittedSearchQueryState.searchCaseSensitivity,
            context.submittedSearchQueryState.searchPatternType,
        ]
    )

    const [allExpanded, setAllExpanded] = useState(false)
    const onExpandAllResultsToggle = useCallback(() => {
        setAllExpanded(oldValue => !oldValue)
        platformContext.telemetryService.log(allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
    }, [allExpanded, platformContext])

    const [showSavedSearchForm, setShowSavedSearchForm] = useState(false)

    // Update local query state on sidebar query state updates.
    useDeepCompareEffectNoCheck(() => {
        if (context.searchSidebarQueryState.proposedQueryState?.queryState) {
            setUserQueryState(context.searchSidebarQueryState.proposedQueryState?.queryState)
        }
    }, [context.searchSidebarQueryState.proposedQueryState?.queryState])

    // Update local search query state on sidebar search submission.
    useDeepCompareEffectNoCheck(() => {
        setUserQueryState(context.submittedSearchQueryState.queryState)
        // It's a whole new object on each state update, so we need
        // to compare (alternatively, construct full query TODO)

        // Clear repo view
        setRepoToShow(null)
    }, [context.submittedSearchQueryState.queryState])

    // Track sidebar + keyboard shortcut search submissions
    useEffect(() => {
        platformContext.telemetryService.log('IDESearchSubmitted')
    }, [platformContext, context.submittedSearchQueryState.queryState.query])

    const onSubmit = useCallback(
        (options?: { caseSensitive?: boolean; patternType?: SearchPatternType; newQuery?: string }) => {
            const previousSearchQueryState = context.submittedSearchQueryState

            const query = options?.newQuery ?? userQueryState.query
            const caseSensitive = options?.caseSensitive ?? previousSearchQueryState.searchCaseSensitivity
            const patternType = options?.patternType ?? previousSearchQueryState.searchPatternType

            extensionCoreAPI
                .streamSearch(query, {
                    caseSensitive,
                    patternType,
                    version: LATEST_VERSION,
                    trace: undefined,
                })
                .then(() => {
                    editorReference.current?.focus()
                })
                .catch(error => {
                    // TODO surface error to users? Errors will typically be caught and
                    // surfaced throught streaming search reuls.
                    console.error(error)
                })

            extensionCoreAPI
                .setSidebarQueryState({
                    queryState: { query },
                    searchCaseSensitivity: caseSensitive,
                    searchPatternType: patternType,
                })
                .catch(error => {
                    // TODO surface error to users
                    console.error('Error updating sidebar query state from panel', error)
                })

            // Log Search History
            const hostname = new URL(instanceURL).hostname
            let queryString = `${userQueryState.query}${caseSensitive ? ' case:yes' : ''}`
            if (context.selectedSearchContextSpec) {
                queryString = appendContextFilter(queryString, context.selectedSearchContextSpec)
            }
            const metrics = queryString ? collectMetrics(queryString) : undefined
            platformContext.telemetryService.log(
                'SearchResultsQueried',
                {
                    code_search: {
                        query_data: {
                            query: metrics,
                            combined: queryString,
                            empty: !queryString,
                        },
                    },
                },
                {
                    code_search: {
                        query_data: {
                            // 🚨 PRIVACY: never provide any private query data in the
                            // { code_search: query_data: query } property,
                            // which is also potentially exported in pings data.
                            query: metrics,

                            // 🚨 PRIVACY: Only collect the full query string for unauthenticated users
                            // on Sourcegraph.com, and only after sanitizing to remove certain filters.
                            combined:
                                !authenticatedUser && isSourcegraphDotCom
                                    ? sanitizeQueryForTelemetry(queryString)
                                    : undefined,
                            empty: !queryString,
                        },
                    },
                },
                `https://${hostname}/search?q=${encodeURIComponent(queryString)}&patternType=${patternType}`
            )
            // Clear repo view
            setRepoToShow(null)
        },
        [
            context.submittedSearchQueryState,
            context.selectedSearchContextSpec,
            userQueryState.query,
            extensionCoreAPI,
            instanceURL,
            platformContext.telemetryService,
            authenticatedUser,
            isSourcegraphDotCom,
        ]
    )

    // Submit new search on change
    const setCaseSensitivity = useCallback(
        (caseSensitive: boolean) => {
            onSubmit({ caseSensitive })
        },
        [onSubmit]
    )

    // Submit new search on change
    const setPatternType = useCallback(
        (patternType: SearchPatternType) => {
            console.log({ patternType })
            onSubmit({ patternType })
        },
        [onSubmit]
    )

    const fetchHighlightedFileLineRangesWithContext = useCallback(
        (parameters: FetchFileParameters) => fetchHighlightedFileLineRanges({ ...parameters, platformContext }),
        [platformContext]
    )

    const fetchStreamSuggestions = useCallback(
        (query: string): Observable<SearchMatch[]> =>
            wrapRemoteObservable(extensionCoreAPI.fetchStreamSuggestions(query, instanceURL)),
        [extensionCoreAPI, instanceURL]
    )

    const globbing = useMemo(() => globbingEnabledFromSettings(settingsCascade), [settingsCascade])

    const setSelectedSearchContextSpec = useCallback(
        (spec: string) => {
            extensionCoreAPI
                .setSelectedSearchContextSpec(spec)
                .catch(error => {
                    console.error('Error persisting search context spec.', error)
                })
                .finally(() => {
                    // Execute search with new context state
                    onSubmit()
                })
        },
        [extensionCoreAPI, onSubmit]
    )

    const onSearchAgain = useCallback(
        (additionalFilters: string[]) => {
            platformContext.telemetryService.log('SearchSkippedResultsAgainClicked')
            onSubmit({
                newQuery: applyAdditionalFilters(context.submittedSearchQueryState.queryState.query, additionalFilters),
            })
        },
        [context.submittedSearchQueryState.queryState, platformContext, onSubmit]
    )

    const onShareResultsClick = useCallback((): void => {
        const queryState = context.submittedSearchQueryState

        const path = `/search?${buildSearchURLQuery(
            queryState.queryState.query,
            queryState.searchPatternType,
            queryState.searchCaseSensitivity,
            context.selectedSearchContextSpec
        )}&utm_campaign=vscode-extension&utm_medium=direct_traffic&utm_source=vscode-extension&utm_content=save-search`
        extensionCoreAPI.copyLink(new URL(path, instanceURL).href).catch(error => {
            console.error('Error copying search link to clipboard:', error)
        })
        platformContext.telemetryService.log('VSCEShareLinkClick')
    }, [context, instanceURL, extensionCoreAPI, platformContext])

    const fullQuery = useMemo(
        () =>
            appendContextFilter(
                context.submittedSearchQueryState.queryState.query ?? '',
                context.selectedSearchContextSpec
            ),
        [context]
    )

    const matchHandlers = useMatchHandlers({
        platformContext,
        extensionCoreAPI,
        authenticatedUser,
        onRepoSelected: setRepoToShow,
        instanceURL,
    })

    const clearRepositoryToShow = (): void => setRepoToShow(null)

    return (
        <div className={styles.resultsViewLayout}>
            {/* eslint-disable-next-line react/forbid-elements */}
            <form
                className="d-flex w-100 my-2 pb-2 border-bottom"
                onSubmit={event => {
                    event.preventDefault()
                    onSubmit()
                }}
            >
                <SearchBox
                    caseSensitive={context.submittedSearchQueryState?.searchCaseSensitivity}
                    setCaseSensitivity={setCaseSensitivity}
                    patternType={context.submittedSearchQueryState?.searchPatternType}
                    setPatternType={setPatternType}
                    isSourcegraphDotCom={isSourcegraphDotCom}
                    structuralSearchDisabled={false}
                    queryState={userQueryState}
                    onChange={onChange}
                    onSubmit={onSubmit}
                    authenticatedUser={authenticatedUser}
                    searchContextsEnabled={true}
                    showSearchContext={true}
                    showSearchContextManagement={false}
                    defaultSearchContextSpec="global"
                    setSelectedSearchContextSpec={setSelectedSearchContextSpec}
                    selectedSearchContextSpec={context.selectedSearchContextSpec}
                    fetchSearchContexts={fetchSearchContexts}
                    fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                    getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                    fetchStreamSuggestions={fetchStreamSuggestions}
                    settingsCascade={settingsCascade}
                    globbing={globbing}
                    isLightTheme={theme === 'theme-light'}
                    telemetryService={platformContext.telemetryService}
                    platformContext={platformContext}
                    className={classNames('flex-grow-1 flex-shrink-past-contents', styles.searchBox)}
                    containerClassName={styles.searchBoxContainer}
                    autoFocus={true}
                    onEditorCreated={setEditor}
                    editorComponent="monaco"
                />
            </form>

            {!repoToShow ? (
                <div className={styles.resultsViewScrollContainer}>
                    <SearchResultsInfoBar
                        onShareResultsClick={onShareResultsClick}
                        showSavedSearchForm={showSavedSearchForm}
                        setShowSavedSearchForm={setShowSavedSearchForm}
                        extensionCoreAPI={extensionCoreAPI}
                        patternType={context.submittedSearchQueryState.searchPatternType}
                        authenticatedUser={authenticatedUser}
                        platformContext={platformContext}
                        stats={
                            <StreamingProgress
                                progress={
                                    context.searchResults?.progress || { durationMs: 0, matchCount: 0, skipped: [] }
                                }
                                state={context.searchResults?.state || 'loading'}
                                onSearchAgain={onSearchAgain}
                                showTrace={false}
                            />
                        }
                        allExpanded={allExpanded}
                        onExpandAllResultsToggle={onExpandAllResultsToggle}
                        instanceURL={instanceURL}
                        fullQuery={fullQuery}
                    />
                    {authenticatedUser && showSavedSearchForm && (
                        <SavedSearchCreateForm
                            authenticatedUser={authenticatedUser}
                            submitLabel="Add saved search"
                            title="Add saved search"
                            fullQuery={`${fullQuery} patternType:${context.submittedSearchQueryState.searchPatternType}`}
                            onComplete={() => setShowSavedSearchForm(false)}
                            platformContext={platformContext}
                            instanceURL={instanceURL}
                        />
                    )}
                    <MatchHandlersContext.Provider value={{ ...matchHandlers, instanceURL }}>
                        <StreamingSearchResultsList
                            isLightTheme={theme === 'theme-light'}
                            settingsCascade={settingsCascade}
                            telemetryService={platformContext.telemetryService}
                            allExpanded={allExpanded}
                            // Debt: dotcom prop used only for "run search" link
                            // for search examples. Fix on VSCE.
                            isSourcegraphDotCom={false}
                            searchContextsEnabled={true}
                            showSearchContext={true}
                            platformContext={platformContext}
                            results={context.searchResults ?? undefined}
                            fetchHighlightedFileLineRanges={fetchHighlightedFileLineRangesWithContext}
                            executedQuery={context.submittedSearchQueryState.queryState.query}
                            resultClassName="mr-0"
                            // TODO "no results" video thumbnail assets
                            // In build, copy ui/assets/img folder to dist/
                            assetsRoot="https://raw.githubusercontent.com/sourcegraph/sourcegraph/main/ui/assets"
                        />
                    </MatchHandlersContext.Provider>
                </div>
            ) : (
                <div className={styles.resultsViewScrollContainer}>
                    <RepoView
                        extensionCoreAPI={extensionCoreAPI}
                        platformContext={platformContext}
                        onBackToSearchResults={clearRepositoryToShow}
                        repositoryMatch={repoToShow}
                        setQueryState={setUserQueryState}
                        instanceURL={instanceURL}
                    />
                </div>
            )}
        </div>
    )
}

const applyAdditionalFilters = (query: string, additionalFilters: string[]): string => {
    let newQuery = query
    for (const filter of additionalFilters) {
        const fieldValue = filter.split(':', 2)
        newQuery = updateFilters(newQuery, fieldValue[0], fieldValue[1])
    }
    return newQuery
}
