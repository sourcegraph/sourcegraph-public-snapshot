/* eslint-disable id-length */
import { Observable, fromEvent, Subscription, OperatorFunction, pipe } from 'rxjs'
import { defaultIfEmpty, scan } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { SearchPatternType } from '../graphql-operations'

// This is an initial proof of concept implementation of search streaming.
// The protocol and implementation is still in the design phase. Feel free to
// change anything and everything here. We are iteratively improving this
// until it is no longer a proof of concept and instead works well.

type SearchEvent = { type: 'filematches'; matches: FileMatch[] } | { type: 'filters'; filters: Filter[] }

interface FileMatch {
    name: string
    repository: string
    branches?: [string]
    version?: string
    lineMatches: [LineMatch]
}

interface LineMatch {
    line: string
    lineNumber: number
    offsetAndLengths: [[number]]
}

interface Filter {
    value: string
    label: string
    count: number
    limitHit: boolean
    kind: string
}

const toGQLLineMatch = (line: LineMatch): GQL.ILineMatch => ({
    __typename: 'LineMatch',
    limitHit: false,
    lineNumber: line.lineNumber,
    offsetAndLengths: line.offsetAndLengths,
    preview: line.line,
})

function toGQLFileMatch(fm: FileMatch): GQL.IFileMatch {
    let revision = ''
    if (fm.branches) {
        const branch = fm.branches[0]
        if (branch !== '') {
            revision = '@' + branch
        }
    } else if (fm.version) {
        revision = '@' + fm.version
    }

    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    const file: GQL.IGitBlob = {
        path: fm.name,
        // /github.com/gorilla/mux@v1.7.2/-/blob/mux_test.go
        // TODO return in response?
        url: '/' + fm.repository + revision + '/-/blob/' + fm.name,
        commit: {
            oid: fm.version || '',
        },
    } as GQL.IGitBlob
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    const repository: GQL.IRepository = {
        name: fm.repository,
    } as GQL.IRepository
    return {
        __typename: 'FileMatch',
        file,
        repository,
        revSpec: null,
        resource: fm.name,
        symbols: [],
        lineMatches: fm.lineMatches.map(toGQLLineMatch),
        limitHit: false,
    }
}

const toGQLSearchFilter = (filter: Omit<Filter, 'type'>): GQL.ISearchFilter => ({
    __typename: 'SearchFilter',
    ...filter,
})

const emptyGQLSearchResults: GQL.ISearchResults = {
    __typename: 'SearchResults',
    matchCount: 0,
    resultCount: 0,
    approximateResultCount: '',
    limitHit: false,
    sparkline: [],
    repositories: [],
    repositoriesCount: 0,
    repositoriesSearched: [],
    indexedRepositoriesSearched: [],
    cloning: [],
    missing: [],
    timedout: [],
    indexUnavailable: false,
    alert: null,
    elapsedMilliseconds: 0,
    dynamicFilters: [],
    results: [],
    pageInfo: { __typename: 'PageInfo', endCursor: null, hasNextPage: false },
}

/**
 * Converts a stream of SearchEvents into an aggregated GQL.ISearchResult
 */
export const switchToGQLISearchResults: OperatorFunction<SearchEvent, GQL.ISearchResults> = pipe(
    scan((results: GQL.ISearchResults, newEvent: SearchEvent) => {
        switch (newEvent.type) {
            case 'filematches':
                return {
                    ...results,
                    // File matches are additive
                    results: results.results.concat(newEvent.matches.map(toGQLFileMatch)),
                }

            case 'filters':
                return {
                    ...results,
                    // New filter results replace all previous ones
                    dynamicFilters: newEvent.filters.map(toGQLSearchFilter),
                }
        }
    }, emptyGQLSearchResults),
    defaultIfEmpty(emptyGQLSearchResults)
)

/**
 * Initiates a streaming search. This is a type safe wrapper around Sourcegraph's streaming search API (using Server Sent Events).
 * The observable will emit each event returned from the backend.
 *
 * @param query the search query to send to Sourcegraph's backend.
 */
export function search(
    query: string,
    version: string,
    patternType: SearchPatternType,
    versionContext: string | undefined
): Observable<SearchEvent> {
    return new Observable<SearchEvent>(observer => {
        const parameters = [
            ['q', query],
            ['v', version],
            ['t', patternType as string],
        ]
        if (versionContext) {
            parameters.push(['vc', versionContext])
        }
        const parameterEncoded = parameters.map(([k, v]) => k + '=' + encodeURIComponent(v)).join('&')

        const eventSource = new EventSource('/search/stream?' + parameterEncoded)
        const subscriptions = new Subscription()
        subscriptions.add(
            fromEvent(eventSource, 'filematches').subscribe((event: Event) => {
                if (!(event instanceof MessageEvent)) {
                    throw new TypeError('internal error: expected MessageEvent in streaming search filematches')
                }
                observer.next({ type: 'filematches', matches: JSON.parse(event.data) as FileMatch[] })
            })
        )
        subscriptions.add(
            fromEvent(eventSource, 'filters').subscribe((event: Event) => {
                if (!(event instanceof MessageEvent)) {
                    throw new TypeError('internal error: expected MessageEvent in streaming search filters')
                }
                observer.next({ type: 'filters', filters: JSON.parse(event.data) as Filter[] })
            })
        )
        subscriptions.add(
            fromEvent(eventSource, 'done').subscribe(() => {
                observer.complete()
                eventSource.close()
            })
        )
        subscriptions.add(
            fromEvent(eventSource, 'error').subscribe(error => {
                observer.error(error)
                eventSource.close()
            })
        )
        return () => {
            subscriptions.unsubscribe()
            eventSource.close()
        }
    })
}
