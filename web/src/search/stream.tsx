/* eslint-disable id-length */
import { Observable, fromEvent, Subscription, OperatorFunction, pipe } from 'rxjs'
import { defaultIfEmpty, map, scan } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { SearchPatternType } from '../graphql-operations'

// This is an initial proof of concept implementation of search streaming.
// The protocol and implementation is still in the design phase. Feel free to
// change anything and everything here. We are iteratively improving this
// until it is no longer a proof of concept and instead works well.

type SearchEvent = FileMatch[]

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

// TODO fill in the fields we actually care about
const toGQLSearchResults = (results: GQL.SearchResult[]): GQL.ISearchResults => ({
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
    results,
    pageInfo: { __typename: 'PageInfo', endCursor: null, hasNextPage: false },
})

/**
 * Converts a stream of SearchEvents into an aggregated GQL.ISearchResult
 */
export const switchToGQLISearchResults: OperatorFunction<SearchEvent, GQL.ISearchResults> = pipe(
    map(fileMatches => fileMatches.map(toGQLFileMatch)),
    scan((allFileMatches: GQL.IFileMatch[], newFileMatches) => allFileMatches.concat(newFileMatches), []),
    defaultIfEmpty([] as GQL.IFileMatch[]),
    map(toGQLSearchResults)
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
                observer.next(JSON.parse(event.data) as FileMatch[])
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
