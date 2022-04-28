import React, { useCallback, useState } from 'react'

import { EMPTY, NEVER, of } from 'rxjs'

import { QueryState, SearchPatternType } from '@sourcegraph/search'
import { SearchBox } from '@sourcegraph/search-ui'
import {
    aggregateStreamingSearch,
    ContentMatch,
    LATEST_VERSION,
    SearchMatch,
} from '@sourcegraph/shared/src/search/stream'
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

export const App: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    onPreviewChange,
    onPreviewClear,
    onOpen,
}: Props) => {
    const [caseSensitive, setCaseSensitivity] = useState(false)
    const [patternType, setPatternType] = useState(SearchPatternType.literal)
    const [results, setResults] = useState<SearchMatch[]>([])
    const [lastSearchedQuery, setLastSearchedQuery] = useState<null | string>(null)
    const [userQueryState, setUserQueryState] = useState<QueryState>({
        query: '',
    })

    const onSubmit = useCallback(() => {
        const query = userQueryState.query

        // When we submit a search that is already the last search, do nothing. This prevents the
        // search results from being reloaded and reapplied in a different order when a user
        // accidently hits enter thinking that this would open the file
        if (query === lastSearchedQuery) {
            return
        }

        aggregateStreamingSearch(of(query), {
            version: LATEST_VERSION,
            patternType,
            caseSensitive,
            trace: undefined,
            sourcegraphURL: 'https://sourcegraph.com/.api',
            decorationContextLines: 0,
            // eslint-disable-next-line rxjs/no-ignored-subscription
        }).subscribe(searchResults => {
            setResults(searchResults.results)
        })
        setResults([])
        setLastSearchedQuery(query)
    }, [caseSensitive, lastSearchedQuery, patternType, userQueryState.query])

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
                            caseSensitive={caseSensitive}
                            setCaseSensitivity={setCaseSensitivity}
                            patternType={patternType}
                            setPatternType={setPatternType}
                            isSourcegraphDotCom={true}
                            hasUserAddedExternalServices={false}
                            hasUserAddedRepositories={true}
                            structuralSearchDisabled={false}
                            queryState={userQueryState}
                            onChange={setUserQueryState}
                            onSubmit={onSubmit}
                            authenticatedUser={null}
                            searchContextsEnabled={false}
                            showSearchContext={true}
                            showSearchContextManagement={false}
                            defaultSearchContextSpec="global"
                            setSelectedSearchContextSpec={setSelectedSearchContextSpec}
                            selectedSearchContextSpec={undefined}
                            fetchSearchContexts={() => {
                                throw new Error('fetchSearchContexts')
                            }}
                            fetchAutoDefinedSearchContexts={() => NEVER}
                            getUserSearchContextNamespaces={() => []}
                            fetchStreamSuggestions={() => NEVER}
                            settingsCascade={EMPTY_SETTINGS_CASCADE}
                            globbing={false}
                            isLightTheme={false}
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                            platformContext={{ requestGraphQL: () => EMPTY }}
                            className=""
                            containerClassName=""
                            autoFocus={true}
                            editorComponent="monaco"
                            hideHelpButton={true}
                        />
                    </form>
                </div>
                <button type="button" onClick={setPrimaryColorToBlack}>Paint it black</button>
                {/* We reset the search result list whenever a new search is started using key={lastSearchedQuery} */}
                <SearchResultList
                    results={results}
                    key={lastSearchedQuery}
                    onPreviewChange={onPreviewChange}
                    onPreviewClear={onPreviewClear}
                    onOpen={onOpen}
                />
            </div>
        </WildcardThemeContext.Provider>
    )
}

function setPrimaryColorToBlack(): void {
    const root = document.querySelector(':root') as HTMLElement
    root.style.setProperty('--primary', 'black')
}

