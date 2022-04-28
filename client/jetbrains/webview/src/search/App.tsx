import React, { useCallback, useState } from 'react'

import { EMPTY, NEVER, of } from 'rxjs'

import { SearchPatternType, QueryState } from '@sourcegraph/search'
import { SearchBox } from '@sourcegraph/search-ui'
import { aggregateStreamingSearch, LATEST_VERSION, SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { WildcardThemeContext } from '@sourcegraph/wildcard'

import styles from './App.module.scss'

export const App: React.FunctionComponent = () => {
    // Toggling case sensitivity or pattern type does NOT trigger a new search on home view.
    const [caseSensitive, setCaseSensitivity] = useState(false)
    const [patternType, setPatternType] = useState(SearchPatternType.literal)
    const [results, setResults] = useState<SearchMatch[]>([])

    const [userQueryState, setUserQueryState] = useState<QueryState>({
        query: '',
    })

    const onSubmit = useCallback(() => {
        aggregateStreamingSearch(of(userQueryState.query), {
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

        console.log('onSubmit')
    }, [caseSensitive, patternType, userQueryState])

    const setSelectedSearchContextSpec = useCallback(() => console.log('setSelectedSearchContextSpec'), [])

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
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
                        searchContextsEnabled={true}
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
                        platformContext={{
                            requestGraphQL: () => EMPTY,
                        }}
                        className=""
                        containerClassName=""
                        autoFocus={true}
                        editorComponent="monaco"
                    />
                </form>
            </div>
            <div>
                <ul>
                    {results.map((match: SearchMatch) =>
                        match.type === 'content'
                            ? match.lineMatches.map(line => (
                                  <li key={`${match.path}-${line.lineNumber}-${JSON.stringify(line.offsetAndLengths)}`}>
                                      {line.line} <small>{match.path}</small>
                                  </li>
                              ))
                            : null
                    )}
                </ul>
            </div>
        </WildcardThemeContext.Provider>
    )
}
