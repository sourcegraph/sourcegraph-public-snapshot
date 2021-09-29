import { isEqual } from 'lodash'
import React, { createContext, Dispatch, SetStateAction, useContext, useEffect, useMemo, useState } from 'react'
import { useHistory } from 'react-router'
import { of } from 'rxjs'
import { throttleTime } from 'rxjs/operators'

import { AggregateStreamingSearchResults, StreamSearchOptions } from '@sourcegraph/shared/src/search/stream'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { SearchStreamingProps } from '..'

interface CachedResults {
    results: AggregateStreamingSearchResults | undefined
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
    options: StreamSearchOptions,
    telemetryService: TelemetryService
): AggregateStreamingSearchResults | undefined {
    const [cachedResults, setCachedResults] = useContext(SearchResultsCacheContext)

    const history = useHistory()

    const results = useObservable(
        useMemo(() => {
            // If options have not changed, return cached value
            if (isEqual(options, cachedResults?.options)) {
                return of(cachedResults?.results)
            }

            return streamSearch(options).pipe(throttleTime(500, undefined, { leading: true, trailing: true }))
        }, [cachedResults?.options, cachedResults?.results, options, streamSearch])
    )

    useEffect(() => {
        if (results?.state === 'complete') {
            setCachedResults({ results, options })
        }
        // Only update cached results if the results change
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [results])

    useEffect(() => {
        // In case of back/forward navigation, log if the cache is being used.
        const cacheExists = isEqual(options, cachedResults?.options)

        if (history.action === 'POP') {
            telemetryService.log('SearchResultsCacheRetrieved', { cacheHit: cacheExists }, { cacheHit: cacheExists })
        }
        // Only log when options have changed
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [options])

    return results
}

export const SearchResultsCacheProvider: React.FunctionComponent<{}> = ({ children }) => {
    const cachedResultsState = useState<CachedResults | null>(null)

    return (
        <SearchResultsCacheContext.Provider value={cachedResultsState}>{children}</SearchResultsCacheContext.Provider>
    )
}
