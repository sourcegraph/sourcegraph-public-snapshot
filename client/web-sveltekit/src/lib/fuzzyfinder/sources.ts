import { Fzf, extendedMatch, byLengthAsc, type FzfOptions, type FzfResultItem } from 'fzf'
import { Observable, Subject } from 'rxjs'
import { throttleTime, switchMap } from 'rxjs/operators'
import { readable, type Readable } from 'svelte/store'

import type { GraphQLClient } from '$lib/graphql'
import { mapOrThrow } from '$lib/graphql'
import type { Loadable } from '$lib/utils'
import { CachedAsyncCompletionSource } from '$lib/web'

import {
    FuzzyFinderQuery,
    type FuzzyFinderSearchResult,
    type FuzzyFinderSearchResult_FileMatch_,
} from './FuzzyFinder.gql'

interface Match {
    /**
     * The value used to match the query.
     */
    value: string
    ranking?: number
}

interface SymbolMatch extends Match {
    type: 'symbol'
    file: FuzzyFinderSearchResult_FileMatch_['file']
    repository: FuzzyFinderSearchResult_FileMatch_['repository']
    symbol: FuzzyFinderSearchResult_FileMatch_['symbols'][number]
}

interface FileMatch extends Match {
    type: 'file'
    file: FuzzyFinderSearchResult_FileMatch_['file']
    repository: FuzzyFinderSearchResult_FileMatch_['repository']
}

interface RepositoryMatch extends Match {
    type: 'repo'
    repository: FuzzyFinderSearchResult_FileMatch_['repository']
}

export type FuzzyFinderResult = SymbolMatch | FileMatch | RepositoryMatch

export interface CompletionSource<T> extends Readable<Loadable<FzfResultItem<T>[]>> {
    next: (value: string) => void
}

// Separate type to enforce that `sort` is always `true` and `tiebreakers` is available.
type FuzzyFinderFzfOptions = FzfOptions<Match> & { sort: true }

// The number of results to fetch from the server and display in the UI.
const LIMIT = 50
const THROTTLE_TIME = 100

const defaultFzfOptions: FuzzyFinderFzfOptions = {
    sort: true,
    fuzzy: 'v2',
    casing: 'smart-case',
    forward: true,
    limit: LIMIT,
    // This enables multi-word matching and special characters like `'`, '^', etc.
    // See https://github.com/junegunn/fzf/tree/7191ebb615f5d6ebbf51d598d8ec853a65e2274d?tab=readme-ov-file#search-syntax
    match: extendedMatch,
    selector: item => item.value,
    tiebreakers: [
        // Sort by ranking, if available.
        (a, b) => (b.item.ranking ?? 0) - (a.item.ranking ?? 0),
    ],
}

export function createRepositorySource(client: GraphQLClient): CompletionSource<RepositoryMatch> {
    return createCompletionSource(
        client,
        () => '',
        value => `type:repo repo:"${value}"`,
        (result, results) => {
            if (result.__typename === 'Repository') {
                results.push([
                    result.name,
                    { type: 'repo', value: result.name, ranking: result.stars, repository: result },
                ])
            }
        },
        {
            forward: false,
        }
    )
}

export function createFileSource(client: GraphQLClient, scope: () => string): CompletionSource<FileMatch> {
    return createCompletionSource(
        client,
        scope,
        (value, scope) => `type:path ${scope} file:"${value}"`,
        (result, results) => {
            if (result.__typename === 'FileMatch') {
                results.push([
                    result.file.url,
                    {
                        type: 'file',
                        value: result.file.path,
                        ranking: result.repository.stars,
                        file: result.file,
                        repository: result.repository,
                    },
                ])
            }
        },
        {
            forward: false,
        }
    )
}

export function createSymbolSource(client: GraphQLClient, scope: () => string): CompletionSource<SymbolMatch> {
    return createCompletionSource(
        client,
        scope,
        (value, scope) => `type:symbol ${scope} "${value}"`,
        (result, results) => {
            if (result.__typename === 'FileMatch') {
                for (const symbol of result.symbols) {
                    results.push([
                        symbol.location.url,
                        {
                            type: 'symbol',
                            value: symbol.name,
                            ranking: result.repository.stars,
                            file: result.file,
                            repository: result.repository,
                            symbol,
                        },
                    ])
                }
            }
        },
        {
            tiebreakers: [byLengthAsc],
        }
    )
}

function createCompletionSource<T extends Match>(
    client: GraphQLClient,
    getScope: () => string,
    getQuery: (value: string, scope?: string) => string,
    appendResult: (result: FuzzyFinderSearchResult, results: [string, T][]) => void,
    additionalFzfOptions?: Partial<FuzzyFinderFzfOptions>
): CompletionSource<T> {
    let fzfOptions = defaultFzfOptions
    if (additionalFzfOptions) {
        fzfOptions = {
            ...fzfOptions,
            ...additionalFzfOptions,
            tiebreakers: (fzfOptions.tiebreakers ?? []).concat(additionalFzfOptions.tiebreakers ?? []),
        }
    }

    const source = new CachedAsyncCompletionSource({
        dataCacheKey: getScope,
        queryKey: (...args) => `count:${LIMIT} ${getQuery(...args)}`,
        async query(query) {
            return client
                .query(FuzzyFinderQuery, {
                    query,
                })
                .then(
                    mapOrThrow(response => {
                        const results: [string, T][] = []
                        for (const result of response.data?.search?.results.results ?? []) {
                            appendResult(result, results)
                        }
                        return results
                    })
                )
        },
        filter(entries, value) {
            const fzf = new Fzf<Match[]>(entries, fzfOptions)
            // @fkling: I wasn't able to get the type inference to work, so I had to cast the result.
            return fzf.find(value) as FzfResultItem<T>[]
        },
    })

    const subject = new Subject<string>()
    const { subscribe } = fromObservable(
        subject.pipe(
            throttleTime(THROTTLE_TIME, undefined, { leading: false, trailing: true }),
            switchMap(value => toObservable(source, value))
        ),
        { pending: true, value: [], error: null }
    )

    return {
        subscribe,
        next: value => subject.next(value),
    }
}

function toObservable<T, U>(source: CachedAsyncCompletionSource<T, U>, value: string): Observable<Loadable<U[]>> {
    return new Observable(subscriber => {
        const result = source.query(value, results => results)
        subscriber.next({ pending: true, value: result.result, error: null })
        result.next().then(result => {
            subscriber.next({ pending: false, value: result.result, error: null })
            subscriber.complete()
        })
    })
}

function fromObservable<T>(observable: Observable<T>, initialValue: T): Readable<T> {
    return readable<T>(initialValue, set => {
        const sub = observable.subscribe(set)
        return () => sub.unsubscribe()
    })
}
