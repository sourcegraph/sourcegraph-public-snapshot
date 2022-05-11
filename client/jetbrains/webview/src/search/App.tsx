import React, { useCallback, useState } from 'react'

import { Observable, of, Subscription } from 'rxjs'

import { requestGraphQLCommon } from '@sourcegraph/http-client'
import {
    fetchAutoDefinedSearchContexts,
    fetchSearchContexts,
    getUserSearchContextNamespaces,
    QueryState,
    SearchPatternType,
} from '@sourcegraph/search'
import { SearchBox } from '@sourcegraph/search-ui'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import {
    aggregateStreamingSearch,
    ContentMatch,
    LATEST_VERSION,
    SearchMatch,
} from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { WildcardThemeContext } from '@sourcegraph/wildcard'

import { SearchResultList } from './results/SearchResultList'

import styles from './App.module.scss'

interface Props {
    onPreviewChange: (match: ContentMatch, lineIndex: number) => void
    onPreviewClear: () => void
    onOpen: (match: ContentMatch, lineIndex: number) => void
}

const requestGraphQL: PlatformContext['requestGraphQL'] = args =>
    requestGraphQLCommon({
        ...args,
        baseUrl: 'https://sourcegraph.com',
    })

function fetchStreamSuggestionsWithStaticUrl(query: string): Observable<SearchMatch[]> {
    return fetchStreamSuggestions(query, 'https://sourcegraph.com/.api')
}

const platformContext = {
    requestGraphQL,
}

export const App: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    onPreviewChange,
    onPreviewClear,
    onOpen,
}: Props) => {
    const [results, setResults] = useState<SearchMatch[]>([])
    const [lastSearchedQuery, setLastSearchedQuery] = useState<{
        query: null | string
        caseSensitive: boolean
        patternType: SearchPatternType
    }>({ query: null, caseSensitive: false, patternType: SearchPatternType.literal })
    const [userQueryState, setUserQueryState] = useState<QueryState>({
        query: '',
    })
    const [subscription, setSubscription] = useState<Subscription>()

    const onSubmit = useCallback(
        (options?: { caseSensitive?: boolean; patternType?: SearchPatternType }) => {
            const query = userQueryState.query
            const caseSensitive = options?.caseSensitive
            const patternType = options?.patternType

            // When we submit a search that is already the last search, do nothing. This prevents the
            // search results from being reloaded and reapplied in a different order when a user
            // accidentally hits enter thinking that this would open the file
            if (
                query === lastSearchedQuery.query &&
                (caseSensitive === undefined || caseSensitive === lastSearchedQuery.caseSensitive) &&
                (patternType === undefined || patternType === lastSearchedQuery.patternType)
            ) {
                return
            }
            // If we don't unsubscribe, the previous search will be continued after the new search and search results will be mixed
            subscription?.unsubscribe()
            setSubscription(
                aggregateStreamingSearch(of(query), {
                    version: LATEST_VERSION,
                    caseSensitive: caseSensitive ?? lastSearchedQuery.caseSensitive,
                    patternType: patternType ?? lastSearchedQuery.patternType,
                    trace: undefined,
                    sourcegraphURL: 'https://sourcegraph.com/.api',
                    decorationContextLines: 0,
                }).subscribe(searchResults => {
                    setResults(searchResults.results)
                })
            )
            setResults([])
            setLastSearchedQuery({
                query,
                caseSensitive: caseSensitive ?? lastSearchedQuery.caseSensitive,
                patternType: patternType ?? lastSearchedQuery.patternType,
            })
        },
        [lastSearchedQuery, subscription, userQueryState.query]
    )

    const setSelectedSearchContextSpec = useCallback(() => console.log('setSelectedSearchContextSpec'), [])

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <div className={styles.root}>
                <div className={styles.searchBoxContainer}>
                    {/* eslint-disable-next-line react/forbid-elements */}
                    <form
                        className="d-flex my-2"
                        onSubmit={event => {
                            event.preventDefault()
                            onSubmit()
                        }}
                    >
                        <SearchBox
                            caseSensitive={lastSearchedQuery.caseSensitive}
                            setCaseSensitivity={caseSensitive => onSubmit({ caseSensitive })}
                            patternType={lastSearchedQuery.patternType}
                            setPatternType={patternType => onSubmit({ patternType })}
                            isSourcegraphDotCom={true} // TODO: Make this dynamic. See VS Code's SearchResultsView.tsx
                            hasUserAddedExternalServices={false}
                            hasUserAddedRepositories={true} // Used for search context CTA, which we won't show here.
                            structuralSearchDisabled={false}
                            queryState={userQueryState}
                            onChange={setUserQueryState}
                            onSubmit={onSubmit}
                            authenticatedUser={null} // TODO: Add authenticated user once we have authentication
                            searchContextsEnabled={true}
                            showSearchContext={true}
                            showSearchContextManagement={false}
                            defaultSearchContextSpec="global"
                            setSelectedSearchContextSpec={setSelectedSearchContextSpec} // TODO: Fix this, see VS Code's SearchResultsView.tsx
                            selectedSearchContextSpec="global" // TODO: Fix this, see VS Code's SearchResultsView.tsx
                            fetchSearchContexts={fetchSearchContexts}
                            fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                            getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                            fetchStreamSuggestions={fetchStreamSuggestionsWithStaticUrl}
                            settingsCascade={EMPTY_SETTINGS_CASCADE} // TODO: Implement this. See VS Code's SearchResultsView.tsx
                            globbing={false} // TODO: Wire it up to plugin settings
                            isLightTheme={false} // TODO: Wire it up with the current theme setting
                            telemetryService={NOOP_TELEMETRY_SERVICE} // TODO: Fix this, see VS Code's SearchResultsView.tsx
                            platformContext={platformContext}
                            className=""
                            containerClassName=""
                            autoFocus={true}
                            editorComponent="monaco"
                            hideHelpButton={true}
                        />
                    </form>
                </div>
                {/* We reset the search result list whenever a new search is started using key={lastSearchedQuery} */}
                <SearchResultList
                    results={results}
                    key={lastSearchedQuery.query}
                    onPreviewChange={onPreviewChange}
                    onPreviewClear={onPreviewClear}
                    onOpen={onOpen}
                />
            </div>
        </WildcardThemeContext.Provider>
    )
}
