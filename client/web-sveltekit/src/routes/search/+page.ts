import { BehaviorSubject, merge, type Observable, of } from 'rxjs'
import { shareReplay } from 'rxjs/operators'
import { get } from 'svelte/store'

import { navigating } from '$app/stores'
import { SearchPatternType } from '$lib/graphql-operations'
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
import { parseSearchURL } from '$lib/web'

import type { PageLoad } from './$types'

const cache: Record<string, Observable<AggregateStreamingSearchResults | undefined>> = {}

export const load: PageLoad = ({ url, depends }) => {
    const hasQuery = url.searchParams.has('q')
    const caseSensitiveURL = url.searchParams.get('case') === 'yes'

    if (hasQuery) {
        let {
            query = '',
            searchMode,
            patternType = SearchPatternType.literal,
            caseSensitive,
        } = parseSearchURL(url.search)
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
        }

        const key = createCacheKey(options, url.search)
        let searchStream = cache[key]

        // Browser back button should always use the cached version if available
        if (get(navigating)?.type !== 'popstate' || !searchStream) {
            const querySource = new BehaviorSubject<string>(query)
            searchStream = cache[key] = merge(of(undefined), aggregateStreamingSearch(querySource, options)).pipe(
                shareReplay(1)
            )
            // Primes the stream
            // eslint-disable-next-line rxjs/no-ignored-subscription
            searchStream.subscribe()
        }
        const resultStream = searchStream
        // Do we actualle need this?
        // merge(searchStream.pipe(throttleTime(500)), searchStream.pipe(last()))

        return {
            stream: resultStream,
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

function createCacheKey(options: StreamSearchOptions, query: string): string {
    return [
        options.version,
        options.patternType,
        options.caseSensitive,
        options.caseSensitive,
        options.searchMode,
        options.chunkMatches,
        query,
    ].join('--')
}
