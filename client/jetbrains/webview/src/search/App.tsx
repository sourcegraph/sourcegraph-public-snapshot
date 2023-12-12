import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { type Observable, of, type Subscription } from 'rxjs'

import { requestGraphQLCommon } from '@sourcegraph/http-client'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import {
    fetchSearchContexts as sharedFetchSearchContexts,
    getUserSearchContextNamespaces,
    type QueryState,
} from '@sourcegraph/shared/src/search'
import {
    aggregateStreamingSearch,
    type AggregateStreamingSearchResults,
    LATEST_VERSION,
    type Progress,
    type SearchMatch,
    type StreamingResultsState,
} from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { EMPTY_SETTINGS_CASCADE, type SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeContext, ThemeSetting } from '@sourcegraph/shared/src/theme'
import { useObservable, WildcardThemeContext } from '@sourcegraph/wildcard'

import { SearchPatternType } from '../graphql-operations'
import { initializeSourcegraphSettings } from '../sourcegraphSettings'

import { getInstanceURL } from '.'
import { fetchSearchContextsCompat } from './compat/fetchSearchContexts'
import { GlobalKeyboardListeners } from './GlobalKeyboardListeners'
import { JetBrainsSearchBox } from './input/JetBrainsSearchBox'
import { saveLastSearch } from './js-to-java-bridge'
import { SearchResultList } from './results/SearchResultList'
import { StatusBar } from './StatusBar'
import type { Search } from './types'

import styles from './App.module.scss'

interface Props {
    isDarkTheme: boolean
    instanceURL: string
    accessToken: string | null
    customRequestHeaders: Record<string, string> | null
    onPreviewChange: (match: SearchMatch, lineOrSymbolMatchIndex?: number) => Promise<void>
    onPreviewClear: () => Promise<void>
    onOpen: (match: SearchMatch, lineOrSymbolMatchIndex?: number) => Promise<void>
    onSearchError: (errorMessage: string) => Promise<void>
    initialSearch: Search | null
    backendVersion: string | null
    authenticatedUser: AuthenticatedUser | null
    telemetryService: TelemetryService
    telemetryRecorder: TelemetryRecorder
}

function fetchStreamSuggestionsWithStaticUrl(query: string): Observable<SearchMatch[]> {
    return fetchStreamSuggestions(query, getInstanceURL() + '.api')
}

const fetchSearchContexts = (backendVersion: string | null): typeof sharedFetchSearchContexts => {
    if (backendVersion === null || backendVersion === '0.0.0+dev') {
        return sharedFetchSearchContexts
    }
    const [major, minor] = backendVersion.split('.').map(Number)
    // This is the case for dotcom
    if (Number.isNaN(major) || Number.isNaN(minor)) {
        return sharedFetchSearchContexts
    }
    // The shared fetchSearchContexts was updated to query additional fields starting from 4.3.
    const isAfterOrOn4_3 = major > 4 || (major === 4 && minor >= 3)
    return isAfterOrOn4_3 ? sharedFetchSearchContexts : fetchSearchContextsCompat
}

function fallbackToLiteralSearchIfNeeded(
    patternType: SearchPatternType | undefined,
    backendVersion: string | null
): SearchPatternType | undefined {
    if (backendVersion === null || patternType !== SearchPatternType.standard) {
        return patternType
    }

    const [major, minor] = backendVersion.split('.').map(Number)
    // SearchPatternType.standard is not supported by versions before 3.43.0
    return major < 3 || (major === 3 && minor < 43) ? SearchPatternType.literal : SearchPatternType.standard
}

export const App: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    isDarkTheme,
    instanceURL,
    accessToken,
    customRequestHeaders,
    onPreviewChange,
    onPreviewClear,
    onOpen,
    onSearchError,
    initialSearch,
    backendVersion,
    authenticatedUser,
    telemetryService,
    telemetryRecorder,
}) => {
    const authState = authenticatedUser !== null ? 'success' : 'failure'

    /**
     * @deprecated Prefer using Apollo-Client instead if possible. The migration is in progress.
     */
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
                    ...customRequestHeaders,
                },
            }),
        [instanceURL, accessToken, customRequestHeaders]
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
            patternType:
                fallbackToLiteralSearchIfNeeded(SearchPatternType.standard, backendVersion) ||
                SearchPatternType.literal,
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
            const patternType = fallbackToLiteralSearchIfNeeded(options?.patternType, backendVersion)
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
                    displayLimit: 200,
                }
            ).subscribe((searchResults: AggregateStreamingSearchResults) => {
                if (searchResults.state === 'error') {
                    setProgressState('error')
                    onSearchError(searchResults.error.message)
                        .then(() => {})
                        .catch(() => {})
                    return
                }
                setMatches(searchResults.results)
                setProgress(searchResults.progress)
                setProgressState(searchResults.state)
            })
            setMatches([])
            setLastSearch(nextSearch)
            saveLastSearch(nextSearch)
            telemetryService.log('IDESearchSubmitted')
            telemetryRecorder.recordEvent('IDESearch', 'submitted')
        },
        [
            lastSearch,
            backendVersion,
            userQueryState.query,
            telemetryService,
            telemetryRecorder,
            instanceURL,
            onSearchError,
        ]
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
                settingsCascade={settingsCascade}
            />
        ),
        [lastSearch, matches, onOpen, onPreviewChange, onPreviewClear, settingsCascade]
    )

    const themeValue = useMemo(
        () => ({ themeSetting: isDarkTheme ? ThemeSetting.Dark : ThemeSetting.Light }),
        [isDarkTheme]
    )

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <ThemeContext.Provider value={themeValue}>
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
                                setSelectedSearchContextSpec={contextSpec => onSubmit({ contextSpec })}
                                selectedSearchContextSpec={lastSearch.selectedSearchContextSpec}
                                fetchSearchContexts={fetchSearchContexts(backendVersion)}
                                getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                                fetchStreamSuggestions={fetchStreamSuggestionsWithStaticUrl}
                                settingsCascade={settingsCascade}
                                telemetryService={telemetryService}
                                telemetryRecorder={telemetryRecorder}
                                platformContext={platformContext}
                                className=""
                                containerClassName=""
                                autoFocus={true}
                                hideHelpButton={true}
                            />
                        </form>
                    </div>

                    {statusBar}

                    {searchResultList}
                </div>
            </ThemeContext.Provider>
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
