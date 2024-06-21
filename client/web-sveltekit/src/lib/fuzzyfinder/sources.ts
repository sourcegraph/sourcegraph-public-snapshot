import { Observable, Subject, from } from 'rxjs'
import { throttleTime, switchMap, startWith } from 'rxjs/operators'
import { readable, type Readable } from 'svelte/store'

import type { GraphQLClient } from '$lib/graphql'
import { mapOrThrow } from '$lib/graphql'
import { scanSearchQueryAsPatterns, stringHuman, PatternKind } from '$lib/shared'
import type { Loadable } from '$lib/utils'

import { FuzzyFinderQuery, type FuzzyFinderFileMatch } from './FuzzyFinder.gql'

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

export interface CompletionSource extends Readable<Loadable<{ results: FuzzyFinderResult[] }>> {
    next: (value: string) => void
}

const QUERY_THROTTLE_TIME = 200

interface FuzzyFinderSourceOptions {
    client: GraphQLClient
    /**
     * Generates the search query given the fuzzy finder input value.
     */
    queryBuilder: (input: string) => string
}

/**
 * Creates a completion source for the fuzzy finder.
 */
export function createFuzzyFinderSource({ client, queryBuilder }: FuzzyFinderSourceOptions): CompletionSource {
    const subject = new Subject<string>()
    const { subscribe } = fromObservable(
        subject.pipe(
            throttleTime(QUERY_THROTTLE_TIME, undefined, { leading: false, trailing: true }),
            switchMap(value =>
                from(
                    client
                        .query(FuzzyFinderQuery, { query: queryBuilder(value) })
                        .then(
                            mapOrThrow(response => {
                                const results: FuzzyFinderResult[] = []
                                for (const result of response?.data?.search?.results.results ?? []) {
                                    switch (result.__typename) {
                                        case 'Repository':
                                            results.push({ type: 'repo', repository: result })
                                            break
                                        case 'FileMatch':
                                            if (result.symbols.length === 0) {
                                                // This is a file match
                                                results.push({
                                                    type: 'file',
                                                    file: result.file,
                                                    repository: result.repository,
                                                })
                                            } else {
                                                // This is a symbol match
                                                for (const symbol of result.symbols) {
                                                    results.push({
                                                        type: 'symbol',
                                                        file: result.file,
                                                        repository: result.repository,
                                                        symbol,
                                                    })
                                                }
                                            }
                                    }
                                }
                                return { results }
                            })
                        )
                        .then(
                            value => ({ pending: false, value, error: null }),
                            error => ({ pending: false, value: { results: [] }, error })
                        )
                ).pipe(startWith({ pending: true, value: { results: [] }, error: null }))
            )
        ),
        { pending: false, value: { results: [] }, error: null }
    )

    return {
        subscribe,
        next: value => subject.next(escapeQuery(value.trim())),
    }
}

/**
 * Converts sepecific token types to normal patterns. E.g. `repo:sourcegraph` should be escaped because
 * we don't want it to be interpreted as a filter.
 *
 * @param query The query to escape.
 * @returns The escaped query.
 */
function escapeQuery(query: string): string {
    const result = scanSearchQueryAsPatterns(query)
    if (result.type !== 'success') {
        return query
    }
    return stringHuman(
        result.term.map(token =>
            token.type === 'pattern' && token.kind === PatternKind.Literal
                ? { ...token, value: `"${escapeQuotes(token.value)}"` }
                : token
        )
    )
}

export const escapeQuery_TEST_ONLY = escapeQuery

function escapeQuotes(value: string): string {
    return value.replaceAll(/"/g, '\\"')
}

function fromObservable<T>(observable: Observable<T>, initialValue: T): Readable<T> {
    return readable<T>(initialValue, set => {
        const sub = observable.subscribe(set)
        return () => sub.unsubscribe()
    })
}
