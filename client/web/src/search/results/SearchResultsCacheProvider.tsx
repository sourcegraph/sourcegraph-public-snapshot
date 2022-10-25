import React, { createContext, Dispatch, SetStateAction, useContext, useEffect, useMemo, useState } from 'react'

import { Remote } from 'comlink'
import { isEqual } from 'lodash'
import { useHistory } from 'react-router'
import { merge, of } from 'rxjs'
import { last, share, throttleTime } from 'rxjs/operators'

import { transformSearchQuery } from '@sourcegraph/shared/src/api/client/search'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { AggregateStreamingSearchResults, StreamSearchOptions } from '@sourcegraph/shared/src/search/stream'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/wildcard'

import { SearchStreamingProps } from '..'
import { useExperimentalFeatures } from '../../stores'
import { eventLogger } from '../../tracking/eventLogger'

interface CachedResults {
    results: AggregateStreamingSearchResults | undefined
    query: string
    options: StreamSearchOptions
}

const SearchResultsCacheContext = createContext<[CachedResults | null, Dispatch<SetStateAction<CachedResults | null>>]>(
    [null, () => null]
)

/**
 * Returns the cached value if the options have not changed.
 * Otherwise, executes a new search and caches the value once
 * the search completes.
 *
 * @param streamSearch Search function.
 * @param options Options to pass on to `streamSeach`. MUST be wrapped in `useMemo` for this to work.
 * @returns Search results, either from cache or from running a new search (updated as new streaming results come in).
 */
export function useCachedSearchResults(
    streamSearch: SearchStreamingProps['streamSearch'],
    query: string,
    options: StreamSearchOptions,
    extensionHostAPI: Promise<Remote<FlatExtensionHostAPI>> | null,
    telemetryService: TelemetryService
): AggregateStreamingSearchResults | undefined {
    const [cachedResults, setCachedResults] = useContext(SearchResultsCacheContext)
    const enableGoImportsSearchQueryTransform = useExperimentalFeatures(
        features => features.enableGoImportsSearchQueryTransform
    )

    const history = useHistory()

    const transformedQuery = useMemo(
        () =>
            transformSearchQuery({
                query,
                extensionHostAPIPromise: extensionHostAPI,
                enableGoImportsSearchQueryTransform,
                eventLogger,
            }),
        [query, extensionHostAPI, enableGoImportsSearchQueryTransform]
    )

    const results = useObservable(
        useMemo(() => {
            // If query and options have not changed, return cached value
            if (query === cachedResults?.query && isEqual(options, cachedResults?.options)) {
                return of(cachedResults?.results)
            }

            const stream = streamSearch(transformedQuery, options).pipe(share())

            // If the throttleTime option `trailing` is set, we will return the
            // final value, but it also removes the guarantee that the output events
            // are a minimum of 200ms apart. If it's unset, we might throw away
            // some trailing events. This is a fundamental issue with throttleTime,
            // and is discussed extensively in github issues. Instead, we just manually
            // merge throttleTime with only leading values and the final value.
            // See: https://github.com/ReactiveX/rxjs/issues/5732
            return merge(stream.pipe(throttleTime(500)), stream.pipe(last()))
        }, [
            query,
            cachedResults?.query,
            cachedResults?.options,
            cachedResults?.results,
            options,
            streamSearch,
            transformedQuery,
        ])
    )

    // Add a history listener that resets cached results if a new search is made
    // with the same query (e.g. to force refresh when the search button is clicked).
    useEffect(() => {
        const unlisten = history.listen((location, action) => {
            if (location.pathname === '/search' && action === 'PUSH') {
                setCachedResults(null)
            }
        })

        return unlisten
    }, [history, setCachedResults])

    useEffect(() => {
        if (results?.state === 'complete') {
            setCachedResults({ results, query, options })
        }
        // Only update cached results if the results change
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [results])

    useEffect(() => {
        // In case of back/forward navigation, log if the cache is being used.
        const cacheExists = query === cachedResults?.query && isEqual(options, cachedResults?.options)

        if (history.action === 'POP') {
            telemetryService.log('SearchResultsCacheRetrieved', { cacheHit: cacheExists }, { cacheHit: cacheExists })
        }
        // Only log when query or options have changed
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [query, options])

    return results
}

export const SearchResultsCacheProvider: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => {
    const cachedResultsState = useState<CachedResults | null>(null)

    return (
        <SearchResultsCacheContext.Provider value={cachedResultsState}>{children}</SearchResultsCacheContext.Provider>
    )
}
