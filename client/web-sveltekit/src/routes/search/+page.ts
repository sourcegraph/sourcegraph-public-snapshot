import { BehaviorSubject, type Observable, of } from 'rxjs'
import { get } from 'svelte/store'

import { browser } from '$app/environment'
import { navigating } from '$app/stores'
import { getGraphQLClient } from '$lib/graphql'
import { SearchPatternType } from '$lib/graphql-operations'
import { parseExtendedSearchURL, type ExtendedParsedSearchURL } from '$lib/search'
import { SearchCachePolicy, getCachePolicyFromURL } from '$lib/search/state'
import {
    aggregateStreamingSearch,
    LATEST_VERSION,
    type AggregateStreamingSearchResults,
    type StreamSearchOptions,
    emptyAggregateResults,
} from '$lib/shared'

import type { PageLoad } from './$types'
import DotcomFooterLinks from './DotcomFooterLinks.svelte'
import { DefaultSearchContext } from './page.gql'
import { queryExampleDotcom, queryExampleEnterprise } from './queryExamples'
import { SearchJob } from './searchJob'

type SearchStreamCacheEntry = Observable<AggregateStreamingSearchResults>

/**
 * CachingStreamManager helps caching and canceling search streams in the browser.
 */
class CachingStreamManager {
    private cache: Map<string, SearchStreamCacheEntry> = new Map()
    private streamManager = new NonCachingStreamManager()

    search(
        parsedQuery: ExtendedParsedSearchURL,
        searchOptions: StreamSearchOptions,
        useCache: boolean
    ): Observable<AggregateStreamingSearchResults> {
        const key = this.createCacheKey(parsedQuery, searchOptions)

        const searchStream = this.cache.get(key)

        if (!useCache || !searchStream) {
            const stream = this.streamManager.search(parsedQuery, searchOptions)
            const searchStream = new BehaviorSubject<AggregateStreamingSearchResults>(emptyAggregateResults)
            this.cache.set(key, searchStream)
            stream.subscribe({
                next: value => {
                    searchStream.next(value)
                },
            })
            return searchStream
        }

        return searchStream
    }

    private createCacheKey(parsedQuery: ExtendedParsedSearchURL, options: StreamSearchOptions): string {
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
}

/**
 * NonCachingStreamManager simply executes the search query without caching.
 */
class NonCachingStreamManager {
    search(
        parsedQuery: ExtendedParsedSearchURL,
        searchOptions: StreamSearchOptions
    ): Observable<AggregateStreamingSearchResults> {
        return aggregateStreamingSearch(of(parsedQuery.filteredQuery ?? ''), searchOptions)
    }
}

const streamManager = browser ? new CachingStreamManager() : new NonCachingStreamManager()

export const load: PageLoad = async ({ parent, url, depends }) => {
    const hasQuery = url.searchParams.has('q')
    const cachePolicy = getCachePolicyFromURL(url)
    const trace = url.searchParams.get('trace') ?? undefined
    const sourcegraphDotComMode = window.context.sourcegraphDotComMode

    const codyHref = sourcegraphDotComMode ? 'https://sourcegraph.com/cody' : '/cody'
    const footer = sourcegraphDotComMode ? DotcomFooterLinks : null

    if (hasQuery) {
        const parsedQuery = parseExtendedSearchURL(url)
        let {
            query = '',
            searchMode,
            patternType = SearchPatternType.keyword,
            caseSensitive,
            filters: queryFilters,
        } = parsedQuery
        depends(`search:${url}`)

        const options: StreamSearchOptions = {
            version: LATEST_VERSION,
            patternType,
            caseSensitive,
            trace,
            // TODO(@camdencheek): populate these from local storage
            featureOverrides: [],
            chunkMatches: true,
            searchMode,
            displayLimit: (await parent()).settings?.['search.displayLimit'] ?? 1500,
            // 5kb is a conservative upper bound on a reasonable line to show
            // to a user. In practice we can likely go much lower.
            maxLineLen: 5 * 1024,
        }

        let useClientCache = false
        switch (cachePolicy) {
            case SearchCachePolicy.CacheFirst:
                useClientCache = true
                break
            case SearchCachePolicy.Default:
                useClientCache = get(navigating)?.type === 'popstate'
                break
        }

        // We create a new stream only if
        // - we do not have a cached stream (in the browser)
        // - the search result page was expliclty navigated to (not via back/forward buttons)
        // - cache is not enforced (which is used in the filters sidebar)
        const searchStream = streamManager.search(parsedQuery, options, useClientCache)

        return {
            codyHref,
            footer,
            searchStream,
            queryFilters,
            queryFromURL: query,
            queryOptions: {
                query,
                caseSensitive,
                patternType,
                searchMode,
            },
            searchJob: window.context.searchJobsEnabled
                ? new SearchJob(getGraphQLClient(), `${query} patternType:${patternType}`)
                : undefined,
        }
    }

    const defaultSearchContext = await getGraphQLClient()
        .query(DefaultSearchContext, {})
        .toPromise()
        .then(result => result.data?.defaultSearchContext?.spec ?? 'global')
        .catch(() => 'global')
    return {
        codyHref,
        footer,
        queryExample: sourcegraphDotComMode ? queryExampleDotcom() : queryExampleEnterprise(),
        showExampleQueries: sourcegraphDotComMode,
        queryOptions: {
            query: `context:${defaultSearchContext} `,
        },
    }
}
