import React, { useCallback, useEffect, useMemo, useState } from 'react'

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
    isDarkTheme: boolean
    instanceURL: string
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

interface Search {
    query: string | null
    caseSensitive: boolean
    patternType: SearchPatternType
    selectedSearchContextSpec: string
}

export const App: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    isDarkTheme,
    instanceURL,
    onPreviewChange,
    onPreviewClear,
    onOpen,
}: Props) => {
    const [results, setResults] = useState<SearchMatch[]>([])
    const [lastSearch, setLastSearch] = useState<Search>({
        query: '',
        caseSensitive: false,
        patternType: SearchPatternType.literal,
        selectedSearchContextSpec: 'global',
    })
    const [userQueryState, setUserQueryState] = useState<QueryState>({
        query: '',
    })
    const [subscription, setSubscription] = useState<Subscription>()

    const isSourcegraphDotCom = useMemo(() => {
        const hostname = new URL(instanceURL).hostname
        return hostname === 'sourcegraph.com' || hostname === 'www.sourcegraph.com'
    }, [instanceURL])

    const onSubmit = useCallback(
        (options?: { caseSensitive?: boolean; patternType?: SearchPatternType; contextSpec?: string }) => {
            const query = userQueryState.query ?? ''
            const caseSensitive = options?.caseSensitive
            const patternType = options?.patternType
            const contextSpec = options?.contextSpec

            // When we submit a search that is already the last search, do nothing. This prevents the
            // search results from being reloaded and reapplied in a different order when a user
            // accidentally hits enter thinking that this would open the file
            if (
                query === lastSearch.query &&
                (caseSensitive === undefined || caseSensitive === lastSearch.caseSensitive) &&
                (patternType === undefined || patternType === lastSearch.patternType) &&
                (contextSpec === undefined || contextSpec === lastSearch.selectedSearchContextSpec)
            ) {
                return
            }
            // If we don't unsubscribe, the previous search will be continued after the new search and search results will be mixed
            subscription?.unsubscribe()
            setSubscription(
                aggregateStreamingSearch(
                    of(`context:${contextSpec ?? lastSearch.selectedSearchContextSpec} ${query}`),
                    {
                        version: LATEST_VERSION,
                        caseSensitive: caseSensitive ?? lastSearch.caseSensitive,
                        patternType: patternType ?? lastSearch.patternType,
                        trace: undefined,
                        sourcegraphURL: 'https://sourcegraph.com/.api',
                        decorationContextLines: 0,
                    }
                ).subscribe(searchResults => {
                    setResults(searchResults.results)
                })
            )
            setResults([])
            setLastSearch(current => ({
                query,
                caseSensitive: caseSensitive ?? current.caseSensitive,
                patternType: patternType ?? current.patternType,
                selectedSearchContextSpec: options?.contextSpec ?? current.selectedSearchContextSpec,
            }))
        },
        [lastSearch, subscription, userQueryState.query]
    )

    useEffect(() => {
        window
            .callJava({ action: 'loadLastSearch', arguments: {} })
            .then(lastSavedSearch => {
                console.log(`Loaded last search: ${JSON.stringify(lastSavedSearch)}`)
                setLastSearch(lastSavedSearch as Search)
            })
            .catch(error => {
                console.error(`Failed to load last search: ${(error as Error).message}`)
            })
    }, [])

    useEffect(() => {
        window
            .callJava({ action: 'saveLastSearch', arguments: lastSearch })
            .then(() => {
                console.log(`Saved last search: ${JSON.stringify(lastSearch)}`)
            })
            .catch(error => {
                console.error(`Failed to save last search: ${(error as Error).message}`)
            })
    }, [lastSearch, userQueryState])

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
                            caseSensitive={lastSearch.caseSensitive}
                            setCaseSensitivity={caseSensitive => onSubmit({ caseSensitive })}
                            patternType={lastSearch.patternType}
                            setPatternType={patternType => onSubmit({ patternType })}
                            isSourcegraphDotCom={isSourcegraphDotCom}
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
                            setSelectedSearchContextSpec={contextSpec => onSubmit({ contextSpec })}
                            selectedSearchContextSpec={lastSearch.selectedSearchContextSpec}
                            fetchSearchContexts={fetchSearchContexts}
                            fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                            getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                            fetchStreamSuggestions={fetchStreamSuggestionsWithStaticUrl}
                            settingsCascade={EMPTY_SETTINGS_CASCADE} // TODO: Implement this. See VS Code's SearchResultsView.tsx
                            globbing={false} // TODO: Wire it up to plugin settings
                            isLightTheme={!isDarkTheme}
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
                    key={lastSearch.query}
                    onPreviewChange={onPreviewChange}
                    onPreviewClear={onPreviewClear}
                    onOpen={onOpen}
                />
            </div>
        </WildcardThemeContext.Provider>
    )
}
