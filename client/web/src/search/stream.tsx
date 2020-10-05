/* eslint-disable id-length */
import { Observable, fromEvent, Subscription, OperatorFunction, pipe } from 'rxjs'
import { defaultIfEmpty, map, scan } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { SearchPatternType } from '../graphql-operations'

// This is an initial proof of concept implementation of search streaming.
// The protocol and implementation is still in the design phase. Feel free to
// change anything and everything here. We are iteratively improving this
// until it is no longer a proof of concept and instead works well.

type SearchEvent =
    | { type: 'filematches'; matches: FileMatch[] }
    | { type: 'repomatches'; matches: RepositoryMatch[] }
    | { type: 'filters'; filters: Filter[] }

interface FileMatch extends RepositoryMatch {
    name: string
    repository: string
    branches?: string[]
    version?: string
    lineMatches: LineMatch[]
}

interface LineMatch {
    line: string
    lineNumber: number
    offsetAndLengths: [[number]]
}

type RepositoryMatch = Pick<FileMatch, 'repository' | 'branches'>

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

// eslint-disable-next-line @typescript-eslint/consistent-type-assertions
const toMarkdown = (text: string): GQL.IMarkdown => ({ __typename: 'Markdown', text } as GQL.IMarkdown)

function toGQLRepositoryMatch(repo: RepositoryMatch): GQL.IRepository {
    const branch = repo?.branches?.[0]
    const revision = branch ? `@${branch}` : ''
    const label = repo.repository + revision

    // We only need to return the subset defined in IGenericSearchResultInterface
    const gqlRepo: unknown = {
        __typename: 'Repository',
        // copy-pasta from repositories.go :'(
        icon:
            'data:image/svg+xml;base64,PHN2ZyB2ZXJzaW9uPSIxLjEiIGlkPSJMYXllcl8xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIKCSB2aWV3Qm94PSIwIDAgNjQgNjQiIHN0eWxlPSJlbmFibGUtYmFja2dyb3VuZDpuZXcgMCAwIDY0IDY0OyIgeG1sOnNwYWNlPSJwcmVzZXJ2ZSI+Cjx0aXRsZT5JY29ucyA0MDA8L3RpdGxlPgo8Zz4KCTxwYXRoIGQ9Ik0yMywyMi40YzEuMywwLDIuNC0xLjEsMi40LTIuNHMtMS4xLTIuNC0yLjQtMi40Yy0xLjMsMC0yLjQsMS4xLTIuNCwyLjRTMjEuNywyMi40LDIzLDIyLjR6Ii8+Cgk8cGF0aCBkPSJNMzUsMjYuNGMxLjMsMCwyLjQtMS4xLDIuNC0yLjRzLTEuMS0yLjQtMi40LTIuNHMtMi40LDEuMS0yLjQsMi40UzMzLjcsMjYuNCwzNSwyNi40eiIvPgoJPHBhdGggZD0iTTIzLDQyLjRjMS4zLDAsMi40LTEuMSwyLjQtMi40cy0xLjEtMi40LTIuNC0yLjRzLTIuNCwxLjEtMi40LDIuNFMyMS43LDQyLjQsMjMsNDIuNHoiLz4KCTxwYXRoIGQ9Ik01MCwxNmgtMS41Yy0wLjMsMC0wLjUsMC4yLTAuNSwwLjV2MzVjMCwwLjMtMC4yLDAuNS0wLjUsMC41aC0yN2MtMC41LDAtMS0wLjItMS40LTAuNmwtMC42LTAuNmMtMC4xLTAuMS0wLjEtMC4yLTAuMS0wLjQKCQljMC0wLjMsMC4yLTAuNSwwLjUtMC41SDQ0YzEuMSwwLDItMC45LDItMlYxMmMwLTEuMS0wLjktMi0yLTJIMTRjLTEuMSwwLTIsMC45LTIsMnYzNi4zYzAsMS4xLDAuNCwyLjEsMS4yLDIuOGwzLjEsMy4xCgkJYzEuMSwxLjEsMi43LDEuOCw0LjIsMS44SDUwYzEuMSwwLDItMC45LDItMlYxOEM1MiwxNi45LDUxLjEsMTYsNTAsMTZ6IE0xOSwyMGMwLTIuMiwxLjgtNCw0LTRjMS40LDAsMi44LDAuOCwzLjUsMgoJCWMxLjEsMS45LDAuNCw0LjMtMS41LDUuNFYzM2MxLTAuNiwyLjMtMC45LDQtMC45YzEsMCwyLTAuNSwyLjgtMS4zQzMyLjUsMzAsMzMsMjkuMSwzMywyOHYtMC42Yy0xLjItMC43LTItMi0yLTMuNQoJCWMwLTIuMiwxLjgtNCw0LTRjMi4yLDAsNCwxLjgsNCw0YzAsMS41LTAuOCwyLjctMiwzLjVoMGMtMC4xLDIuMS0wLjksNC40LTIuNSw2Yy0xLjYsMS42LTMuNCwyLjQtNS41LDIuNWMtMC44LDAtMS40LDAuMS0xLjksMC4zCgkJYy0wLjIsMC4xLTEsMC44LTEuMiwwLjlDMjYuNiwzOCwyNywzOC45LDI3LDQwYzAsMi4yLTEuOCw0LTQsNHMtNC0xLjgtNC00YzAtMS41LDAuOC0yLjcsMi0zLjRWMjMuNEMxOS44LDIyLjcsMTksMjEuNCwxOSwyMHoiLz4KPC9nPgo8L3N2Zz4K',
        label: toMarkdown(`[${label}](/${label})`),
        url: '/' + label,
        detail: toMarkdown('Repository name match'),
        matches: [],
    }

    return gqlRepo as GQL.IRepository
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

            case 'repomatches':
                return {
                    ...results,
                    // Repository matches are additive
                    results: results.results.concat(newEvent.matches.map(toGQLRepositoryMatch)),
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

const observeMessages = <T extends {}>(eventSource: EventSource, eventName: SearchEvent['type']): Observable<T> =>
    fromEvent(eventSource, eventName).pipe(
        map((event: Event) => {
            if (!(event instanceof MessageEvent)) {
                throw new TypeError(`internal error: expected MessageEvent in streaming search ${eventName}`)
            }
            try {
                const parsedData = JSON.parse(event.data) as T
                return parsedData
            } catch {
                throw new Error(`Could not parse ${eventName} message data in streaming search`)
            }
        })
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
    versionContext: string | undefined,
    graph: string | undefined
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
        if (graph) {
            parameters.push(['g', graph])
        }
        const parameterEncoded = parameters.map(([k, v]) => k + '=' + encodeURIComponent(v)).join('&')

        const eventSource = new EventSource('/search/stream?' + parameterEncoded)
        const subscriptions = new Subscription()
        subscriptions.add(
            observeMessages<FileMatch[]>(eventSource, 'filematches')
                .pipe(map(matches => ({ type: 'filematches' as const, matches })))
                .subscribe(observer)
        )
        subscriptions.add(
            observeMessages<RepositoryMatch[]>(eventSource, 'repomatches')
                .pipe(map(matches => ({ type: 'repomatches' as const, matches })))
                .subscribe(observer)
        )
        subscriptions.add(
            observeMessages<Filter[]>(eventSource, 'filters')
                .pipe(map(filters => ({ type: 'filters' as const, filters })))
                .subscribe(observer)
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
