import classNames from 'classnames'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { SearchBox } from '@sourcegraph/branded/src/search/input/SearchBox'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'
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

import { SearchResult, SearchVariables } from '../../graphql-operations'
import { WebviewPageProps } from '../platform/context'

import styles from './index.module.scss'
import { searchQuery } from './queries'
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

    const instanceHostname = useMemo(() => sourcegraphVSCodeExtensionAPI.getInstanceHostname(), [
        sourcegraphVSCodeExtensionAPI,
    ])

    const sourcegraphSettings =
        useObservable(
            useMemo(() => wrapRemoteObservable(sourcegraphVSCodeExtensionAPI.getSettings()), [
                sourcegraphVSCodeExtensionAPI,
            ])
        ) ?? EMPTY_SETTINGS_CASCADE

    const globbing = useMemo(() => globbingEnabledFromSettings(sourcegraphSettings), [sourcegraphSettings])

    const [loading, setLoading] = useState(false)

    const onSubmit = useCallback(() => {
        searchActions.submitQuery()
    }, [searchActions])

    useEffect(() => {
        const subscriptions = new Subscription()

        if (queryToRun.query) {
            setLoading(true)

            let queryString = `${queryToRun.query}${caseSensitive ? ' case:yes' : ''}`

            if (selectedSearchContextSpec) {
                queryString = appendContextFilter(queryString, selectedSearchContextSpec)
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
    }, [queryToRun, patternType, caseSensitive, selectedSearchContextSpec, searchActions, platformContext])

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
    }

    return (
        <div>
            <div className="d-flex my-2">
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
            </div>
            {loading ? (
                <p>Loading...</p>
            ) : (
                <SearchResults
                    platformContext={platformContext}
                    theme={theme}
                    sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                    settings={sourcegraphSettings}
                    instanceHostname={instanceHostname}
                />
            )}
        </div>
    )
}
