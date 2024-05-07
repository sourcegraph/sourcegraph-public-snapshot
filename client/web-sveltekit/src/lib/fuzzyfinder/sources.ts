import type { GraphQLClient } from '$lib/graphql'
import { CachedAsyncCompletionSource } from '$lib/web'
import { FuzzyFinderQuery, type FuzzyFinderFileMatch } from './FuzzyFinder.gql'
import { mapOrThrow } from '$lib/graphql'
import { Fzf, type FzfOptions, type FzfResultItem } from 'fzf'
import { readable, type Readable } from 'svelte/store'
import type { Loadable } from '$lib/utils'
import { Observable, Subject } from 'rxjs'
import { throttleTime, switchMap } from 'rxjs/operators'

interface SymbolMatch {
    type: 'symbol'
    file: FuzzyFinderFileMatch['file']
    repository: FuzzyFinderFileMatch['repository']
    symbol: FuzzyFinderFileMatch['symbols'][number]
}

interface FileMatch {
    type: 'file'
    file: FuzzyFinderFileMatch['file']
    repository: FuzzyFinderFileMatch['repository']
}

interface RepositoryMatch {
    type: 'repo'
    repository: FuzzyFinderFileMatch['repository']
}

export type FuzzyFinderResult = SymbolMatch | FileMatch | RepositoryMatch

export interface CompletionSource<T> extends Readable<Loadable<FzfResultItem<T>[]>> {
    next: (value: string) => void
}

export function createRepositorySource(client: GraphQLClient): CompletionSource<RepositoryMatch> {
    const fzfOptions: FzfOptions<RepositoryMatch> = {
        sort: true,
        fuzzy: 'v2',
        selector: item => item.repository.name,
        forward: false,
        limit: 50,
        tiebreakers: [(a, b) => b.item.repository.stars - a.item.repository.stars, (a, b) => b.start - a.start],
    }

    const source = new CachedAsyncCompletionSource({
        queryKey(value) {
            return `type:repo count:50 repo:"${value}"`
        },
        async query(query) {
            return client
                .query(FuzzyFinderQuery, {
                    query,
                })
                .then(
                    mapOrThrow(response => {
                        const repos: [string, RepositoryMatch][] = []
                        for (const result of response.data?.search?.results.results ?? []) {
                            if (result.__typename === 'Repository') {
                                repos.push([result.name, { type: 'repo', repository: result }])
                            }
                        }
                        return repos
                    })
                )
        },
        filter(entries, value) {
            return new Fzf(entries, fzfOptions).find(value)
        },
    })

    const subject = new Subject<string>()
    const { subscribe } = fromObservable(
        subject.pipe(
            throttleTime(100, undefined, { leading: false, trailing: true }),
            switchMap(value => toObservable(source, value))
        ),
        { pending: true, value: [], error: null }
    )

    return {
        subscribe,
        next: value => subject.next(value),
    }
}

export function createFileSource(client: GraphQLClient, scope: () => string): CompletionSource<FileMatch> {
    const fzfOptions: FzfOptions<FileMatch> = {
        sort: true,
        fuzzy: 'v2',
        selector: item => item.file.path,
        forward: false,
        limit: 50,
        tiebreakers: [(a, b) => b.item.repository.stars - a.item.repository.stars, (a, b) => b.start - a.start],
    }

    const source = new CachedAsyncCompletionSource({
        dataCacheKey: scope,
        queryKey(value, scope) {
            return `type:path count:50 ${scope} file:"${value}"`
        },
        async query(query) {
            return client
                .query(FuzzyFinderQuery, {
                    query,
                })
                .then(
                    mapOrThrow(response => {
                        const repos: [string, FileMatch][] = []
                        for (const result of response.data?.search?.results.results ?? []) {
                            if (result.__typename === 'FileMatch') {
                                repos.push([
                                    result.file.url,
                                    { type: 'file', file: result.file, repository: result.repository },
                                ])
                            }
                        }
                        return repos
                    })
                )
        },
        filter(entries, value) {
            return new Fzf(entries, fzfOptions).find(value)
        },
    })

    const subject = new Subject<string>()
    const { subscribe } = fromObservable(
        subject.pipe(
            throttleTime(100, undefined, { leading: false, trailing: true }),
            switchMap(value => toObservable(source, value))
        ),
        { pending: true, value: [], error: null }
    )

    return {
        subscribe,
        next: value => subject.next(value),
    }
}

export function createSymbolSource(client: GraphQLClient, scope: () => string): CompletionSource<SymbolMatch> {
    const fzfOptions: FzfOptions<SymbolMatch> = {
        sort: true,
        fuzzy: 'v2',
        selector: item => item.symbol.name,
        limit: 50,
        tiebreakers: [(a, b) => b.item.repository.stars - a.item.repository.stars, (a, b) => b.start - a.start],
    }

    const source = new CachedAsyncCompletionSource({
        dataCacheKey: scope,
        queryKey(value, scope) {
            return `type:symbol count:50 ${scope} "${value}"`
        },
        async query(query) {
            return client
                .query(FuzzyFinderQuery, {
                    query,
                })
                .then(
                    mapOrThrow(response => {
                        const results: [string, SymbolMatch][] = []
                        for (const result of response.data?.search?.results.results ?? []) {
                            if (result.__typename === 'FileMatch') {
                                for (const symbol of result.symbols) {
                                    results.push([
                                        symbol.location.url,
                                        { type: 'symbol', file: result.file, repository: result.repository, symbol },
                                    ])
                                }
                            }
                        }
                        return results
                    })
                )
        },
        filter(entries, value) {
            return new Fzf(entries, fzfOptions).find(value)
        },
    })

    const subject = new Subject<string>()
    const { subscribe } = fromObservable(
        subject.pipe(
            throttleTime(100, undefined, { leading: false, trailing: true }),
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
