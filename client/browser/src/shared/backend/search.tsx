import { flatten } from 'lodash'
import { Observable, Subject } from 'rxjs'
import {
    catchError,
    debounceTime,
    distinctUntilChanged,
    filter,
    map,
    mergeMap,
    publishReplay,
    refCount,
    repeat,
    switchMap,
    take,
    toArray,
} from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../../shared/src/platform/context'
import { isPrivateRepository } from '../util/context'
import { createAggregateError } from './errors'

interface BaseSuggestion {
    title: string
    description?: string

    /**
     * The URL that is navigated to when the user selects this
     * suggestion.
     */
    url: string

    /**
     * A label describing the action taken when navigating to
     * the URL (e.g., "go to repository").
     */
    urlLabel: string
}

interface SymbolSuggestion extends BaseSuggestion {
    type: 'symbol'
    kind: string
}

interface RepoSuggestion extends BaseSuggestion {
    type: 'repo'
}

interface FileSuggestion extends BaseSuggestion {
    type: 'file'
}

interface DirSuggestion extends BaseSuggestion {
    type: 'dir'
}

export type Suggestion = SymbolSuggestion | RepoSuggestion | FileSuggestion | DirSuggestion

/**
 * Returns all but the last element of path, or "." if that would be the empty path.
 */
function dirname(path: string): string | undefined {
    return (
        path
            .split('/')
            .slice(0, -1)
            .join('/') || '.'
    )
}

/**
 * Returns the last element of path, or "." if path is empty.
 */
function basename(path: string): string {
    return path.split('/').slice(-1)[0] || '.'
}

function createSuggestion(item: GQL.SearchSuggestion): Suggestion | null {
    switch (item.__typename) {
        case 'Repository': {
            return {
                type: 'repo',
                title: item.name,
                url: item.url,
                urlLabel: 'go to repository',
            }
        }
        case 'File': {
            const descriptionParts: string[] = []
            const dir = dirname(item.path)
            if (dir !== undefined && dir !== '.') {
                descriptionParts.push(`${dir}/`)
            }
            descriptionParts.push(basename(item.repository.name))
            if (item.isDirectory) {
                return {
                    type: 'dir',
                    title: item.name,
                    description: descriptionParts.join(' — '),
                    url: item.url,
                    urlLabel: 'go to dir',
                }
            }
            return {
                type: 'file',
                title: item.name,
                description: descriptionParts.join(' — '),
                url: item.url,
                urlLabel: 'go to file',
            }
        }
        case 'Symbol': {
            return {
                type: 'symbol',
                kind: item.kind,
                title: item.name,
                description: `${item.containerName || item.location.resource.path} — ${basename(
                    item.location.resource.repository.name
                )}`,
                url: item.url,
                urlLabel: 'go to definition',
            }
        }
        default:
            return null
    }
}

const symbolsFragment = gql`
    fragment SymbolFields on Symbol {
        __typename
        name
        containerName
        url
        kind
        location {
            resource {
                path
                repository {
                    name
                }
            }
            url
        }
    }
`

const fetchSuggestions = (query: string, first: number, queryGraphQL: PlatformContext['requestGraphQL']) =>
    queryGraphQL<GQL.IQuery>(
        gql`
            query SearchSuggestions($query: String!, $first: Int!) {
                search(query: $query) {
                    suggestions(first: $first) {
                        ... on Repository {
                            __typename
                            name
                            url
                        }
                        ... on File {
                            __typename
                            path
                            name
                            isDirectory
                            url
                            repository {
                                name
                            }
                        }
                        ... on Symbol {
                            ...SymbolFields
                        }
                    }
                }
            }
            ${symbolsFragment}
        `,
        {
            query,
            // The browser extension API only takes 5 suggestions
            first,
        },
        // This request may contain private info if the repository is private
        isPrivateRepository()
    ).pipe(
        map(dataOrThrowErrors),
        mergeMap(({ search }) => {
            if (!search || !search.suggestions) {
                throw new Error('No search suggestions')
            }
            return search.suggestions
        })
    )

interface SuggestionInput {
    query: string
    handler: (suggestion: Suggestion[]) => void
}

export const createSuggestionFetcher = (first = 5, queryGraphQL: PlatformContext['requestGraphQL']) => {
    const fetcher = new Subject<SuggestionInput>()

    fetcher
        .pipe(
            distinctUntilChanged(),
            debounceTime(200),
            switchMap(({ query, handler }) =>
                fetchSuggestions(query, first, queryGraphQL).pipe(
                    take(first),
                    map(createSuggestion),
                    // createSuggestion will return null if we get a type we don't recognize
                    filter((f): f is Suggestion => !!f),
                    toArray(),
                    map((suggestions: Suggestion[]) => ({
                        suggestions,
                        suggestHandler: handler,
                    })),
                    publishReplay(),
                    refCount()
                )
            ),
            // But resubscribe afterwards
            repeat()
        )
        .subscribe(({ suggestions, suggestHandler }) => suggestHandler(suggestions))

    return (input: SuggestionInput) => fetcher.next(input)
}

export const fetchSymbols = (
    query: string,
    queryGraphQL: PlatformContext['requestGraphQL']
): Observable<GQL.ISymbol[]> =>
    queryGraphQL<GQL.IQuery>(
        gql`
            query SearchResults($query: String!) {
                search(query: $query) {
                    results {
                        results {
                            ... on FileMatch {
                                symbols {
                                    ...SymbolFields
                                }
                            }
                        }
                    }
                }
            }
            ${symbolsFragment}
        `,
        {
            query,
        },
        // This request may contain private info if the repository is private
        isPrivateRepository()
    ).pipe(
        map(dataOrThrowErrors),
        map(({ search }) => {
            if (!search) {
                throw new Error('fetchSymbols: empty search')
            }
            if (!search.results) {
                throw new Error('fetchSymbols: empty search.results')
            }

            const symbolsResults = flatten((search.results.results as GQL.IFileMatch[]).map(match => match.symbols))

            return symbolsResults
        }),
        catchError(err => {
            // TODO@ggilmore: This is a kludge that should be removed once the
            // code smells with requestGraphQL are addressed.
            // At this time of writing, requestGraphQL throws the entire response
            // instead of a well-formed error created from response.errors. This kludge
            // manually creates this well-formed error before re-throwing it.
            //
            // See https://github.com/sourcegraph/browser-extension/pull/235 for more context.

            if (err.errors) {
                throw createAggregateError(err.errors)
            }

            throw err
        })
    )
