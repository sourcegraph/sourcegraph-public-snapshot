import { BehaviorSubject, type Observable, of, type Subscription } from 'rxjs'
import { get } from 'svelte/store'

import { browser } from '$app/environment'
import { navigating } from '$app/stores'
import { SearchPatternType } from '$lib/graphql-operations'
import { parseExtendedSearchURL, type ExtendedParsedSearchURL } from '$lib/search'
import { USE_CLIENT_CACHE_QUERY_PARAMETER } from '$lib/search/constants'
import {
    aggregateStreamingSearch,
    LATEST_VERSION,
    type AggregateStreamingSearchResults,
    type StreamSearchOptions,
    filterExists,
    FilterType,
    getGlobalSearchContextFilter,
    omitFilter,
} from '$lib/shared'

import type { PageLoad } from './$types'

interface SearchStreamCacheEntry {
    searchStream: Observable<AggregateStreamingSearchResults | undefined>
    complete: boolean
}

/**
 * CachingStreamManager helps caching and canceling search streams in the browser.
 */
class CachingStreamManager {
    private cache: Map<string, SearchStreamCacheEntry> = new Map()
    private activeSubscription: { cacheKey: string; subscription: Subscription } | undefined
    private streamManager = new NonCachingStreamManager()

    search(
        parsedQuery: ExtendedParsedSearchURL,
        searchOptions: StreamSearchOptions,
        bypassCache: boolean
    ): Observable<AggregateStreamingSearchResults | undefined> {
        const key = createCacheKey(parsedQuery, searchOptions)

        // Cancel any active query to reduce load on the server
        {
            const { cacheKey, subscription } = this.activeSubscription ?? {}
            if (cacheKey && cacheKey !== key) {
                subscription?.unsubscribe()
                // We need to remove the cache entry for ongoing queries otherwise
                // the cache will contain partial data
                if (!this.cache.get(cacheKey)?.complete) {
                    this.cache.delete(cacheKey)
                }
            }
        }

        const searchStream = this.cache.get(key)?.searchStream

        if (bypassCache || !searchStream) {
            const stream = this.streamManager.search(parsedQuery, searchOptions)
            const searchStream = new BehaviorSubject<AggregateStreamingSearchResults | undefined>(undefined)
            const cacheEntry: SearchStreamCacheEntry = { searchStream, complete: false }
            this.cache.set(key, cacheEntry)
            // Primes the stream
            const subscription = stream.subscribe({
                next: value => {
                    searchStream.next(value)
                },
                complete: () => {
                    cacheEntry.complete = true
                },
            })
            this.activeSubscription = { cacheKey: key, subscription }
            return searchStream
        }

        return searchStream
    }
}

/**
 * NonCachingStreamManager simply executes the search query without caching.
 */
class NonCachingStreamManager {
    search(
        parsedQuery: ExtendedParsedSearchURL,
        searchOptions: StreamSearchOptions
    ): Observable<AggregateStreamingSearchResults | undefined> {
        return aggregateStreamingSearch(of(parsedQuery.filteredQuery ?? ''), searchOptions)
    }
}

const streamManager = browser ? new CachingStreamManager() : new NonCachingStreamManager()

export const load: PageLoad = ({ url, depends }) => {
    const hasQuery = url.searchParams.has('q')
    const caseSensitiveURL = url.searchParams.get('case') === 'yes'
    const forceCache = url.searchParams.has(USE_CLIENT_CACHE_QUERY_PARAMETER)

    if (hasQuery) {
        const parsedQuery = parseExtendedSearchURL(url)
        let {
            query = '',
            searchMode,
            patternType = SearchPatternType.keyword,
            caseSensitive,
            filters: queryFilters,
        } = parsedQuery
        // Necessary for allowing to submit the same query again
        // FIXME: This is not correct
        depends(`query:${query}--${caseSensitiveURL}`)

        let searchContext = 'global'
        if (filterExists(query, FilterType.context)) {
            // TODO: Validate search context
            const globalSearchContext = getGlobalSearchContextFilter(query)
            if (globalSearchContext?.spec) {
                searchContext = globalSearchContext.spec
                query = omitFilter(query, globalSearchContext.filter)
            }
        }

        const options: StreamSearchOptions = {
            version: LATEST_VERSION,
            patternType,
            caseSensitive,
            trace: '',
            featureOverrides: [],
            chunkMatches: true,
            searchMode,
            displayLimit: 500,
            maxLineLen: 1000,
        }

        // We create a new stream only if
        // - we do not have a cached stream (in the browser)
        // - the search result page was expliclty navigated to (not via back/forward buttons)
        // - cache is not enforced (which is used in the filters sidebar)
        const searchStream = streamManager.search(
            parsedQuery,
            options,
            !forceCache && get(navigating)?.type !== 'popstate'
        )

        return {
            searchStream,
            queryFilters,
            queryOptions: {
                query,
                caseSensitive,
                patternType,
                searchMode,
                searchContext,
            },
        }
    }
    return {
        queryOptions: {
            query: '',
        },
    }
}

function createCacheKey(parsedQuery: ExtendedParsedSearchURL, options: StreamSearchOptions): string {
    return [
        options.version,
        options.patternType,
        options.caseSensitive,
        options.searchMode,
        options.chunkMatches,
        parsedQuery.filteredQuery,
        parsedQuery.searchMode,
        parsedQuery.patternType,
        parsedQuery.caseSensitive,
    ].join('--')
}
