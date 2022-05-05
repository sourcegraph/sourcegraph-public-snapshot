import React, { useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'
import { Observable } from 'rxjs'
import { useDeepCompareEffectNoCheck } from 'use-deep-compare-effect'

import {
    SearchPatternType,
    getUserSearchContextNamespaces,
    fetchAutoDefinedSearchContexts,
    QueryState,
} from '@sourcegraph/search'
import { SearchBox } from '@sourcegraph/search-ui'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { collectMetrics } from '@sourcegraph/shared/src/search/query/metrics'
import { appendContextFilter, sanitizeQueryForTelemetry } from '@sourcegraph/shared/src/search/query/transformer'
import { LATEST_VERSION, SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { globbingEnabledFromSettings } from '@sourcegraph/shared/src/util/globbing'

import { SearchHomeState } from '../../state'
import { WebviewPageProps } from '../platform/context'

import { fetchSearchContexts } from './alias/fetchSearchContext'
import { BrandHeader } from './components/BrandHeader'
import { HomeFooter } from './components/HomeFooter'

import styles from './index.module.scss'
export interface SearchHomeViewProps extends WebviewPageProps {
    context: SearchHomeState['context']
}

export const SearchHomeView: React.FunctionComponent<React.PropsWithChildren<SearchHomeViewProps>> = ({
    extensionCoreAPI,
    authenticatedUser,
    platformContext,
    settingsCascade,
    theme,
    context,
    instanceURL,
}) => {
    // Toggling case sensitivity or pattern type does NOT trigger a new search on home view.
    const [caseSensitive, setCaseSensitivity] = useState(false)
    const [patternType, setPatternType] = useState(SearchPatternType.literal)

    const [userQueryState, setUserQueryState] = useState<QueryState>({
        query: '',
    })

    const isSourcegraphDotCom = useMemo(() => {
        const hostname = new URL(instanceURL).hostname
        return hostname === 'sourcegraph.com' || hostname === 'www.sourcegraph.com'
    }, [instanceURL])

    const onSubmit = useCallback(() => {
        extensionCoreAPI
            .streamSearch(userQueryState.query, {
                caseSensitive,
                patternType,
                version: LATEST_VERSION,
                trace: undefined,
            })
            .catch(error => {
                // TODO surface error to users? Errors will typically be caught and
                // surfaced throught streaming search reuls.
                console.error(error)
            })

        extensionCoreAPI
            .setSidebarQueryState({
                queryState: { query: userQueryState.query },
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
    }, [
        extensionCoreAPI,
        userQueryState.query,
        caseSensitive,
        patternType,
        instanceURL,
        context.selectedSearchContextSpec,
        platformContext.telemetryService,
        authenticatedUser,
        isSourcegraphDotCom,
    ])

    // Update local query state on sidebar query state updates.
    useDeepCompareEffectNoCheck(() => {
        if (context.searchSidebarQueryState.proposedQueryState?.queryState) {
            setUserQueryState(context.searchSidebarQueryState.proposedQueryState?.queryState)
        }
    }, [context.searchSidebarQueryState.proposedQueryState?.queryState])

    const globbing = useMemo(() => globbingEnabledFromSettings(settingsCascade), [settingsCascade])

    const setSelectedSearchContextSpec = useCallback(
        (spec: string) => {
            extensionCoreAPI.setSelectedSearchContextSpec(spec).catch(error => {
                console.error('Error persisting search context spec.', error)
            })
        },
        [extensionCoreAPI]
    )

    const fetchStreamSuggestions = useCallback(
        (query): Observable<SearchMatch[]> =>
            wrapRemoteObservable(extensionCoreAPI.fetchStreamSuggestions(query, instanceURL)),
        [extensionCoreAPI, instanceURL]
    )

    return (
        <div className="d-flex flex-column align-items-center">
            <BrandHeader theme={theme} />

            <div className={styles.homeSearchBoxContainer}>
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
                        isSourcegraphDotCom={isSourcegraphDotCom}
                        hasUserAddedExternalServices={false}
                        hasUserAddedRepositories={true} // Used for search context CTA, which we won't show here.
                        structuralSearchDisabled={false}
                        queryState={userQueryState}
                        onChange={setUserQueryState}
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
                        editorComponent="monaco"
                    />
                </form>

                <HomeFooter
                    isLightTheme={theme === 'theme-light'}
                    setQuery={setUserQueryState}
                    telemetryService={platformContext.telemetryService}
                />
            </div>
        </div>
    )
}
