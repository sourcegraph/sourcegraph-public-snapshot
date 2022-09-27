import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { Observable, of, Subscription } from 'rxjs'

import { requestGraphQLCommon } from '@sourcegraph/http-client'
import {
    fetchAutoDefinedSearchContexts,
    fetchSearchContexts,
    getUserSearchContextNamespaces,
    QueryState,
    SearchPatternType,
} from '@sourcegraph/search'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import {
    aggregateStreamingSearch,
    LATEST_VERSION,
    Progress,
    SearchMatch,
    StreamingResultsState,
} from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable, WildcardThemeContext } from '@sourcegraph/wildcard'

import { initializeSourcegraphSettings } from '../sourcegraphSettings'

import { GlobalKeyboardListeners } from './GlobalKeyboardListeners'
import { JetBrainsSearchBox } from './input/JetBrainsSearchBox'
import { saveLastSearch } from './js-to-java-bridge'
import { SearchResultList } from './results/SearchResultList'
import { StatusBar } from './StatusBar'
import { Search } from './types'

import { getInstanceURL } from '.'

import styles from './App.module.scss'

interface Props {
    isDarkTheme: boolean
    instanceURL: string
    isGlobbingEnabled: boolean
    accessToken: string | null
    onPreviewChange: (match: SearchMatch, lineOrSymbolMatchIndex?: number) => Promise<void>
    onPreviewClear: () => Promise<void>
    onOpen: (match: SearchMatch, lineOrSymbolMatchIndex?: number) => Promise<void>
    initialSearch: Search | null
    authenticatedUser: AuthenticatedUser | null
    telemetryService: TelemetryService
}

function fetchStreamSuggestionsWithStaticUrl(query: string): Observable<SearchMatch[]> {
    return fetchStreamSuggestions(query, getInstanceURL() + '.api')
}

export const App: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    isDarkTheme,
    instanceURL,
    isGlobbingEnabled,
    accessToken,
    onPreviewChange,
    onPreviewClear,
    onOpen,
    initialSearch,
    authenticatedUser,
    telemetryService,
}: Props) => {
    const authState = authenticatedUser !== null ? 'success' : 'failure'

    const requestGraphQL = useCallback<PlatformContext['requestGraphQL']>(
        args =>
            requestGraphQLCommon({
                ...args,
                baseUrl: instanceURL,
                headers: {
                    Accept: 'application/json',
                    'Content-Type': 'application/json',
                    'X-Sourcegraph-Should-Trace': new URLSearchParams(window.location.search).get('trace') || 'false',
                    ...(accessToken && { Authorization: `token ${accessToken}` }),
                },
            }),
        [instanceURL, accessToken]
    )

    const settingsCascade: SettingsCascadeOrError =
        useObservable(useMemo(() => initializeSourcegraphSettings(requestGraphQL).settings, [requestGraphQL])) ||
        EMPTY_SETTINGS_CASCADE

    const platformContext = {
        requestGraphQL,
    }

    const [matches, setMatches] = useState<SearchMatch[]>([])
    const [progress, setProgress] = useState<Progress>({ durationMs: 0, matchCount: 0, skipped: [] })
    const [progressState, setProgressState] = useState<StreamingResultsState | null>(null)
    const [lastSearch, setLastSearch] = useState<Search>(
        initialSearch ?? {
            query: '',
            caseSensitive: false,
            patternType: SearchPatternType.standard,
            selectedSearchContextSpec: 'global',
        }
    )
    const [userQueryState, setUserQueryState] = useState<QueryState>({
        query: lastSearch.query ?? '',
    })
    const subscription = useRef<Subscription>()

    const isSourcegraphDotCom = useMemo(() => {
        const hostname = new URL(instanceURL).hostname
        return hostname === 'sourcegraph.com' || hostname === 'www.sourcegraph.com'
    }, [instanceURL])

    const onSubmit = useCallback(
        (options?: {
            caseSensitive?: boolean
            patternType?: SearchPatternType
            contextSpec?: string
            forceNewSearch?: true
        }) => {
            const query = userQueryState.query ?? ''
            const caseSensitive = options?.caseSensitive
            const patternType = options?.patternType
            const contextSpec = options?.contextSpec
            const forceNewSearch = options?.forceNewSearch ?? false

            // When we submit a search that is already the last search, do nothing. This prevents the
            // search results from being reloaded and reapplied in a different order when a user
            // accidentally hits enter thinking that this would open the file
            if (
                !forceNewSearch &&
                query === lastSearch.query &&
                (caseSensitive === undefined || caseSensitive === lastSearch.caseSensitive) &&
                (patternType === undefined || patternType === lastSearch.patternType) &&
                (contextSpec === undefined || contextSpec === lastSearch.selectedSearchContextSpec)
            ) {
                return
            }

            const nextSearch = {
                query,
                caseSensitive: caseSensitive ?? lastSearch.caseSensitive,
                patternType: patternType ?? lastSearch.patternType,
                selectedSearchContextSpec: options?.contextSpec ?? lastSearch.selectedSearchContextSpec,
            }

            // If we don't unsubscribe, the previous search will be continued after the new search and search results will be mixed
            subscription.current?.unsubscribe()
            subscription.current = aggregateStreamingSearch(
                of(`context:${nextSearch.selectedSearchContextSpec} ${query}`),
                {
                    version: LATEST_VERSION,
                    caseSensitive: nextSearch.caseSensitive,
                    patternType: nextSearch.patternType,
                    trace: undefined,
                    sourcegraphURL: instanceURL + '.api',
                    decorationContextLines: 0,
                    displayLimit: 200,
                }
            ).subscribe(searchResults => {
                setMatches(searchResults.results)
                setProgress(searchResults.progress)
                setProgressState(searchResults.state)
            })
            setMatches([])
            setLastSearch(nextSearch)
            saveLastSearch(nextSearch)
            telemetryService.log('IDESearchSubmitted')
        },
        [lastSearch, userQueryState.query, telemetryService, instanceURL]
    )

    const [didInitialSubmit, setDidInitialSubmit] = useState(false)
    useEffect(() => {
        if (didInitialSubmit) {
            return
        }
        setDidInitialSubmit(true)

        if (initialSearch !== null) {
            onSubmit({
                caseSensitive: initialSearch.caseSensitive,
                patternType: initialSearch.patternType,
                contextSpec: initialSearch.selectedSearchContextSpec,
                forceNewSearch: true,
            })
        }
    }, [initialSearch, onSubmit, didInitialSubmit])

    const statusBar = useMemo(
        () => <StatusBar progress={progress} progressState={progressState} authState={authState} />,
        [progress, progressState, authState]
    )

    // We reset the search result list whenever a new search is initiated using key={getStableKeyForLastSearch(lastSearch)}
    const searchResultList = useMemo(
        () => (
            <SearchResultList
                matches={matches}
                key={getStableKeyForLastSearch(lastSearch)}
                onPreviewChange={onPreviewChange}
                onPreviewClear={onPreviewClear}
                onOpen={onOpen}
            />
        ),
        [lastSearch, matches, onOpen, onPreviewChange, onPreviewClear]
    )

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <GlobalKeyboardListeners />
            {/* eslint-disable-next-line jsx-a11y/no-static-element-interactions */}
            <div className={styles.root} onMouseDown={preventAll}>
                <div className={styles.searchBoxContainer}>
                    {/* eslint-disable-next-line react/forbid-elements */}
                    <form
                        className="d-flex m-0"
                        onSubmit={event => {
                            event.preventDefault()
                            onSubmit()
                        }}
                    >
                        <JetBrainsSearchBox
                            caseSensitive={lastSearch.caseSensitive}
                            setCaseSensitivity={caseSensitive => onSubmit({ caseSensitive })}
                            patternType={lastSearch.patternType}
                            setPatternType={patternType => onSubmit({ patternType })}
                            isSourcegraphDotCom={isSourcegraphDotCom}
                            structuralSearchDisabled={false}
                            queryState={userQueryState}
                            onChange={setUserQueryState}
                            onSubmit={onSubmit}
                            authenticatedUser={authenticatedUser}
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
                            settingsCascade={settingsCascade}
                            globbing={isGlobbingEnabled}
                            isLightTheme={!isDarkTheme}
                            telemetryService={telemetryService}
                            platformContext={platformContext}
                            className=""
                            containerClassName=""
                            autoFocus={true}
                            editorComponent="monaco"
                            hideHelpButton={true}
                        />
                    </form>
                </div>

                {statusBar}

                {searchResultList}
            </div>
        </WildcardThemeContext.Provider>
    )
}

function getStableKeyForLastSearch(lastSearch: Search): string {
    return `${lastSearch.query ?? ''}-${lastSearch.caseSensitive}-${String(lastSearch.patternType)}-${
        lastSearch.selectedSearchContextSpec
    }`
}

function preventAll(event: React.MouseEvent): void {
    event.stopPropagation()
    event.preventDefault()
}
