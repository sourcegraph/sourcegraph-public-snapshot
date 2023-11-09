import type { EditorState } from '@codemirror/state'
import { mdiFilterOutline, mdiSourceRepository, mdiStar, mdiFileOutline } from '@mdi/js'
import { byLengthAsc, extendedMatch, Fzf, type FzfOptions, type FzfResultItem } from 'fzf'

// This module implements suggestions for the experimental search input

// eslint-disable-next-line no-restricted-imports
import {
    type Group,
    type Option,
    type Source,
    type SuggestionResult,
    combineResults,
    defaultLanguages,
    RenderAs,
} from '@sourcegraph/branded/src/search-ui/experimental'
import { getQueryInformation } from '@sourcegraph/branded/src/search-ui/input/codemirror/parsedQuery'
import { gql } from '@sourcegraph/http-client'
import { getUserSearchContextNamespaces } from '@sourcegraph/shared/src/search'
import { getRelevantTokens } from '@sourcegraph/shared/src/search/query/analyze'
import { regexInsertText } from '@sourcegraph/shared/src/search/query/completion-utils'
import {
    FILTERS,
    FilterType,
    isNegatableFilter,
    type ResolvedFilter,
} from '@sourcegraph/shared/src/search/query/filters'
import type { Node, Parameter } from '@sourcegraph/shared/src/search/query/parser'
import { predicateCompletion } from '@sourcegraph/shared/src/search/query/predicates'
import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { selectorHasFields } from '@sourcegraph/shared/src/search/query/selectFilter'
import { type CharacterRange, type Filter, KeywordKind, type Token } from '@sourcegraph/shared/src/search/query/token'
import { isFilterOfType, resolveFilterMemoized } from '@sourcegraph/shared/src/search/query/utils'
import { getSymbolIconSVGPath } from '@sourcegraph/shared/src/symbols/symbolIcons'

import type { AuthenticatedUser } from '../../auth'
import type {
    SuggestionsRepoResult,
    SuggestionsRepoVariables,
    SuggestionsFileResult,
    SuggestionsFileVariables,
    SuggestionsSymbolResult,
    SuggestionsSymbolVariables,
    SymbolKind,
    SuggestionsSearchContextResult,
    SuggestionsSearchContextVariables,
} from '../../graphql-operations'
import { CachedAsyncCompletionSource } from '../autocompletion/source'

// The number of entries we want to show in various situations
//
// The number of filter values to show when there are multiple sections (e.g. values and predicates)
const MULTIPLE_FILTER_VALUE_LIST_SIZE = 7
// The number of filter values to show when there is only one section
const ALL_FILTER_VALUE_LIST_SIZE = 12
// The number of default suggestions
const DEFAULT_SUGGESTIONS_LIST_SIZE = 3
// The number of default suggestions for important types
const DEFAULT_SUGGESTIONS_HIGH_PRI_LIST_SIZE = 5

/**
 * Used to organize the various sources that contribute to the final list of
 * suggestions.
 */
type InternalSource<T extends Token | undefined = Token | undefined> = (params: {
    token: T
    tokens: Token[]
    parsedQuery: Node | null
    input: string
    position: number
}) => SuggestionResult | null

const none: any[] = []

function starTiebraker(a: { item: { stars: number } }, b: { item: { stars: number } }): number {
    return b.item.stars - a.item.stars
}

/**
 * Ranks default and starred contexts higher than others
 */
function contextTiebraker(a: { item: Context }, b: { item: Context }): number {
    return (b.item.starred || b.item.default ? 1 : 0) - (a.item.starred || a.item.default ? 1 : 0)
}

// `id` is used as cache key
const REPOS_QUERY = gql`
    query SuggestionsRepo($query: String!) {
        search(patternType: regexp, query: $query) {
            results {
                repositories {
                    id
                    name
                    stars
                }
            }
        }
    }
`

// `canonicalURL` is used as cache key
const FILE_QUERY = gql`
    query SuggestionsFile($query: String!) {
        search(patternType: regexp, query: $query) {
            results {
                results {
                    ... on FileMatch {
                        __typename
                        file {
                            path
                            url
                            canonicalURL
                            repository {
                                name
                                stars
                            }
                        }
                    }
                }
            }
        }
    }
`

// `canonicalURL` is used as cache key
const SYMBOL_QUERY = gql`
    query SuggestionsSymbol($query: String!) {
        search(patternType: regexp, query: $query) {
            results {
                results {
                    ... on FileMatch {
                        __typename
                        file {
                            path
                            canonicalURL
                        }
                        symbols {
                            kind
                            url
                            canonicalURL
                            name
                        }
                    }
                }
            }
        }
    }
`

const SEARCH_CONTEXT_QUERY = gql`
    query SuggestionsSearchContext($first: Int!, $query: String, $namespaces: [ID]) {
        searchContexts(
            first: $first
            query: $query
            namespaces: $namespaces
            after: null
            orderBy: SEARCH_CONTEXT_SPEC
            descending: false
        ) {
            nodes {
                id
                name
                spec
                description
                viewerHasStarred
                viewerHasAsDefault
            }
        }
    }
`

interface Repo {
    name: string
    stars: number
}

interface Context {
    name: string
    spec: string
    default: boolean
    starred: boolean
    description: string
}

interface File {
    path: string
    // The repository stars
    stars: number
    repository: string
    url: string
}

interface CodeSymbol {
    kind: SymbolKind
    name: string
    url: string
    path: string
}

/**
 * Converts a Repo value to a suggestion.
 */
function toRepoSuggestion(result: FzfResultItem<Repo>, from: number, to?: number): Option {
    const option = toRepoCompletion(result, from, to, 'repo:')
    option.render = RenderAs.FILTER
    return option
}

/**
 * Converts a Repo value to a completion suggestion.
 */
function toRepoCompletion(
    { item, positions }: FzfResultItem<Repo>,
    from: number,
    to?: number,
    valuePrefix = ''
): Option {
    if (valuePrefix) {
        positions = shiftPositions(positions, valuePrefix.length)
    }
    return {
        label: valuePrefix + item.name,
        matches: positions,
        icon: mdiSourceRepository,
        kind: 'repo',
        action: {
            type: 'completion',
            insertValue: valuePrefix + regexInsertText(item.name) + ' ',
            from,
            to,
        },
        alternativeAction: {
            type: 'goto',
            url: `/${item.name}`,
        },
    }
}

/**
 * Converts a Context value to a completion suggestion.
 */
function toContextCompletion({ item, positions }: FzfResultItem<Context>, from: number, to?: number): Option {
    let description = item.default ? 'Default' : ''
    if (item.description) {
        if (item.default) {
            description += '・'
        }
        description += item.description
    }

    return {
        label: item.spec,
        // Passing an empty string is a hack to draw an "empty" icon
        icon: item.starred ? mdiStar : ' ',
        description,
        matches: positions,
        kind: 'context',
        action: {
            type: 'completion',
            insertValue: item.spec + ' ',
            from,
            to,
        },
    }
}

/**
 * Converts a filter to a completion suggestion.
 */
function toFilterCompletion(label: string, description: string | undefined, from: number, to?: number): Option {
    return {
        label,
        icon: mdiFilterOutline,
        render: RenderAs.FILTER,
        description,
        kind: 'filter',
        action: {
            type: 'completion',
            insertValue: label + ':',
            from,
            to,
        },
    }
}

/**
 * Converts a File value to a completion suggestion.
 */
function toFileCompletion(
    { item, positions }: FzfResultItem<File>,
    from: number,
    to?: number,
    valuePrefix = ''
): Option {
    if (valuePrefix) {
        positions = shiftPositions(positions, valuePrefix.length)
    }
    return {
        label: valuePrefix + item.path,
        icon: mdiFileOutline,
        matches: positions,
        kind: 'file',
        action: {
            type: 'completion',
            insertValue: valuePrefix + regexInsertText(item.path) + ' ',
            from,
            to,
        },
    }
}

/**
 * Converts a File value to a (jump) target suggestion.
 */
function toFileSuggestion(result: FzfResultItem<File>, from: number, to?: number): Option {
    const option = toFileCompletion(result, from, to, 'file:')
    option.render = RenderAs.FILTER
    return option
}

/**
 * Converts a File value to a (jump) target suggestion.
 */
function toSymbolSuggestion(
    { item, positions }: FzfResultItem<CodeSymbol>,
    includeType: boolean,
    from: number,
    to?: number
): Option {
    return {
        label: item.name,
        matches: positions,
        icon: getSymbolIconSVGPath(item.kind),
        kind: 'symbol',
        action: {
            type: 'completion',
            insertValue: (includeType ? 'type:symbol ' : '') + item.name + ' ',
            from,
            to,
        },
    }
}

const FILTER_MATCHER = new Fzf(Object.keys(FILTERS) as FilterType[], { match: extendedMatch })
const NEGATEABLE_FILTER_MATCHER = new Fzf(
    Object.keys(FILTERS).filter(filterType => isNegatableFilter(filterType as FilterType)),
    { match: extendedMatch }
)
// These are the filters shown when the query input is empty or the cursor is at
// at whitespace token.
const DEFAULT_FILTERS: FilterType[] = [
    FilterType.repo,
    FilterType.lang,
    FilterType.type,
    FilterType.select,
    FilterType.context,
]
// If the query contains one of the listed filters, suggest these filters
// too.
const RELATED_FILTERS: Partial<Record<FilterType, (filter: Filter) => FilterType[]>> = {
    [FilterType.type]: filter => {
        switch (filter.value?.value) {
            case 'diff':
            case 'commit': {
                return [FilterType.author, FilterType.before, FilterType.after, FilterType.message]
            }
        }
        return []
    },
    [FilterType.repo]: () => [FilterType.file],
}

/**
 * If the input is empty or the cursor is at a whitespace token, show default suggestions.
 */
const defaultSuggestions: InternalSource = ({ tokens, token, position }) => {
    let options: Group['options'] = []

    if (token && token.type !== 'whitespace') {
        return null
    }

    const isEmpty = tokens.length === 0

    const filters = DEFAULT_FILTERS
        // Show related filters
        .concat(
            tokens.flatMap(token => {
                if (token.type !== 'filter') {
                    return none
                }
                const resolvedFilter = resolveFilterMemoized(token.field.value)
                return resolvedFilter ? RELATED_FILTERS[resolvedFilter.type]?.(token) ?? none : none
            })
        )
        // Remove existing filters
        .filter(filterType => !tokens.some(token => token.type === 'filter' && isFilterOfType(token, filterType)))

    options = filters.map(filter =>
        toFilterCompletion(filter, getFilterDescription(resolveFilterMemoized(filter)), position)
    )

    if (
        (token?.type === 'whitespace' || (!token && !isEmpty)) &&
        !tokens.some(token => token.type === 'keyword' && token.kind === KeywordKind.Or)
    ) {
        options.push({
            label: 'OR',
            description: 'Matches the left or the right side',
            render: RenderAs.QUERY,
            kind: 'keyword',
            icon: ' ', // for alignment
            action: {
                type: 'completion',
                insertValue: 'OR ',
                from: position,
            },
        })
    }

    return options.length > 0 ? { result: [{ title: 'Narrow your search', options }] } : null
}

/**
 * Returns filter completion suggestions for the current term at the cursor.
 * Filters are matched by prefix.
 */
const filterSuggestions: InternalSource = ({ token }) => {
    if (token?.type !== 'pattern') {
        return null
    }

    let options: Group['options'] = []

    if (token.value.startsWith('-')) {
        options = NEGATEABLE_FILTER_MATCHER.find('^' + token.value.slice(1)).map(entry => {
            const resolvedFilter = resolveFilterMemoized(entry.item)
            return {
                ...toFilterCompletion(
                    '-' + entry.item,
                    getFilterDescription(resolvedFilter, true),
                    token.range.start,
                    token.range.end
                ),
                matches: shiftPositions(entry.positions, 1),
            }
        })
    } else {
        // ^ triggers a prefix match
        options = FILTER_MATCHER.find('^' + token.value).flatMap(entry => {
            const resolvedFilter = resolveFilterMemoized(entry.item)
            const options = [
                {
                    ...toFilterCompletion(
                        entry.item,
                        getFilterDescription(resolvedFilter),
                        token.range.start,
                        token.range.end
                    ),
                    matches: entry.positions,
                },
            ]
            if (resolvedFilter && isNegatableFilter(resolvedFilter?.type)) {
                options.push({
                    ...toFilterCompletion(
                        '-' + entry.item,
                        getFilterDescription(resolvedFilter, true),
                        token.range.start,
                        token.range.end
                    ),
                    matches: shiftPositions(entry.positions, 1),
                })
            }
            return options
        })
    }

    return options.length > 0 ? { result: [{ title: 'Narrow your search', options }] } : null
}

const contextActions: Group = {
    title: 'Actions',
    options: [
        {
            label: 'Manage contexts',
            description: 'Add, edit, remove search contexts',
            kind: 'command',
            action: {
                type: 'goto',
                name: 'Go to /contexts',
                url: '/contexts',
            },
        },
    ],
}

/**
 * Returns static and dynamic completion suggestions for filters when completing
 * a filter value.
 */
function filterValueSuggestions(caches: Caches): InternalSource {
    return ({ token, parsedQuery, position }) => {
        if (token?.type !== 'filter') {
            return null
        }
        const resolvedFilter = resolveFilterMemoized(token.field.value)

        if (!resolvedFilter) {
            return null
        }

        const value = token.value?.value ?? ''
        // The value is always inserted after the filter field
        const from = token.value?.range.start ?? token.range.end
        const to = token.value?.range.end ?? token.range.end

        switch (resolvedFilter.definition.suggestions) {
            case 'repo': {
                const predicates = staticFilterPredicateOptions('repo', token, from, to)
                return caches.repo.query(
                    value,
                    entries => {
                        const groups = [
                            {
                                title: 'Repositories',
                                options: entries
                                    .slice(
                                        0,
                                        predicates.length === 0
                                            ? ALL_FILTER_VALUE_LIST_SIZE
                                            : MULTIPLE_FILTER_VALUE_LIST_SIZE
                                    )
                                    .map(item => toRepoCompletion(item, from, to)),
                            },
                        ]

                        if (predicates.length > 0) {
                            groups.push({
                                title: 'Predicates',
                                options: predicates,
                            })
                        }

                        return groups
                    },
                    parsedQuery,
                    position
                )
            }

            case 'path': {
                const predicates = staticFilterPredicateOptions('file', token, from, to)
                return caches.file.query(
                    value,
                    entries => {
                        const groups = [
                            {
                                title: 'Files',
                                options: limitUniqueOptions(
                                    entries,
                                    predicates.length === 0
                                        ? ALL_FILTER_VALUE_LIST_SIZE
                                        : MULTIPLE_FILTER_VALUE_LIST_SIZE,
                                    item => toFileCompletion(item, from, to)
                                ),
                            },
                        ]

                        if (predicates.length > 0) {
                            groups.push({
                                title: 'Predicates',
                                options: predicates,
                            })
                        }

                        return groups
                    },
                    parsedQuery,
                    position
                )
            }

            default: {
                switch (resolvedFilter.type) {
                    // Some filters are not defined to have dynamic suggestions,
                    // we need to handle these here explicitly. We can't change
                    // the filter definition without breaking the current
                    // search input.
                    case FilterType.context: {
                        return caches.context.query(value, entries => {
                            entries = value.trim() === '' ? entries.slice(0, ALL_FILTER_VALUE_LIST_SIZE) : entries
                            return [
                                {
                                    title: 'Search contexts',
                                    options: entries.map(entry => toContextCompletion(entry, from, to)),
                                },
                                contextActions,
                            ]
                        })
                    }
                    default: {
                        const options = staticFilterValueOptions(token, resolvedFilter)
                        return options.length > 0 ? { result: [{ title: '', options }] } : null
                    }
                }
            }
        }
    }
}

const filterValueFzfOptions: Partial<Record<FilterType, Partial<FzfOptions<Option>>>> = {
    [FilterType.lang]: {
        fuzzy: 'v2',
    },
}

function staticFilterValueOptions(
    token: Extract<Token, { type: 'filter' }>,
    resolvedFilter: NonNullable<ResolvedFilter>
): Option[] {
    if (!resolvedFilter.definition.discreteValues) {
        return []
    }

    const value = token.value?.value ?? ''
    // The value is always inserted after the filter field
    const from = token.value?.range.start ?? token.range.end
    const to = token.value?.range.end ?? token.range.end

    let options: Option[]
    if (resolvedFilter.type === FilterType.select) {
        // The some select filter values have multiple subfields, e.g.
        // "symbol.class". To provide a balanced list of suggestions and
        // ergonomics we show subfields only if the value already contains a
        // "top-level" value (e.g. "symbol" or "commit"). To make this work
        // selecting a top-level value with subfields should _not_ append a space
        // for starting a new token. This is what `selectorHasFields` determines
        // below.
        // At the same time, if we already show all subfields (including the
        // top-level value), then selecting any of the values should also append
        // a space. This is handled by the `includesSubFieldValues` check.
        //
        // Examples:
        // - Selecting "repo" will append "repo " (repo has no subfields)
        // - Selecting "symbol" will append "symbol", which in turn will list
        //   all "symbol" related values (including "symbol" itself)
        // - Selecting any of the "symbol..." values inserts that value
        //   including a trailing space because all of them are "terminal"
        //   values at this point.
        const values = resolvedFilter.definition.discreteValues(token.value, false)
        const includesSubFieldValues = values.some(value => value.label.includes('.'))

        options = values.map(({ label }) => ({
            label,
            kind: 'filter-value-select',
            action: {
                type: 'completion',
                from,
                to,
                insertValue: selectorHasFields(label) && !includesSubFieldValues ? label : label + ' ',
            },
        }))
    } else if (resolvedFilter.type === FilterType.lang && !value) {
        // We show a shorter default languages list than the current query
        // input.
        options = defaultLanguages.map(label => ({
            label,
            kind: 'filter-value-lang',
            action: {
                type: 'completion',
                from,
                to,
                insertValue: label + ' ',
            },
        }))
    } else {
        options = resolvedFilter.definition.discreteValues(token.value, false).map(value => ({
            label: value.label,
            description: value.description,
            kind: `filter-value-${resolvedFilter.type}`,
            action: {
                type: 'completion',
                from,
                to,
                insertValue: (value.insertText ?? value.label) + ' ',
            },
        }))
    }

    if (value) {
        const fzf = new Fzf(options, {
            selector: option => option.label,
            fuzzy: false,
            ...filterValueFzfOptions[resolvedFilter.type],
        })
        options = fzf.find(value).map(match => ({ ...match.item, matches: match.positions }))
    }

    return options
}

type PredicateFzfOptions = FzfOptions<{ label: string; asSnippet?: boolean; insertText?: string }>
const predicateFzfOption: PredicateFzfOptions = {
    selector: completion => completion.label,
    fuzzy: false,
    forward: false,
    tiebreakers: [byStartDesc, byLengthAsc],
}

/**
 * Returns predicate options for the provided filter type.
 */
function staticFilterPredicateOptions(type: 'repo' | 'file', filter: Filter, from: number, to: number): Option[] {
    const fzf = new Fzf(predicateCompletion(type), predicateFzfOption)
    return fzf.find(filter.value?.value || '').map(({ item, positions }) => ({
        label: item.label,
        description: item.description,
        matches: positions,
        kind: `filter-predicate-${type}`,
        action: {
            type: 'completion',
            from,
            to,
            // insertText is always set for prediction completions
            insertValue: item.insertText! + ' ${}',
            asSnippet: item.asSnippet,
        },
    }))
}

/**
 * Returns repository (jump) target suggestions matching the term at the cursor,
 * but only if the query doesn't already contain a 'repo:' filter.
 */
function repoSuggestions(cache: Caches['repo']): InternalSource {
    return ({ token, tokens, parsedQuery, position }) => {
        const showRepoSuggestions =
            token?.type === 'pattern' &&
            !tokens.some(token => token.type === 'filter' && isFilterOfType(token, FilterType.repo))
        if (!showRepoSuggestions) {
            return null
        }

        return cache.query(
            token.value,
            results => [
                {
                    title: 'Repositories',
                    options: results
                        .slice(0, DEFAULT_SUGGESTIONS_LIST_SIZE)
                        .map(result => toRepoSuggestion(result, token.range.start)),
                },
            ],
            parsedQuery,
            position
        )
    }
}

/**
 * Returns file (jump) target suggestions matching the term at the cursor,
 * but only if the query contains suitable filters. On dotcom we only show file
 * suggestions if the query contains at least one context: or repo: filter.
 */
function fileSuggestions(cache: Caches['file'], isSourcegraphDotCom?: boolean): InternalSource {
    return ({ token, tokens, parsedQuery, position }) => {
        // Only show file suggestions on dotcom if the query contains at least
        // one context: filter that is not 'global', or a repo: filter.
        const showFileSuggestions =
            token?.type === 'pattern' &&
            (!isSourcegraphDotCom ||
                tokens.some(token => {
                    if (token.type !== 'filter') {
                        return false
                    }
                    return (
                        (isFilterOfType(token, FilterType.context) && token.value?.value !== 'global') ||
                        isFilterOfType(token, FilterType.repo)
                    )
                }))

        if (!showFileSuggestions) {
            return null
        }

        return cache.query(
            token.value,
            results => [
                {
                    title: 'Files',
                    options: limitUniqueOptions(results, DEFAULT_SUGGESTIONS_HIGH_PRI_LIST_SIZE, result =>
                        toFileSuggestion(result, token.range.start)
                    ),
                },
            ],
            parsedQuery,
            position
        )
    }
}

/**
 * Returns file (jump) target suggestions matching the term at the cursor.
 * Because symbol queries are expensive and are slower the less "precise" the
 * query is we are only showing symbol suggestions if the query contains suitable
 * filters (context, repo or file; context must be different from global).
 */
function symbolSuggestions(cache: Caches['symbol']): InternalSource {
    return ({ token, tokens, parsedQuery, position }) => {
        if (token?.type !== 'pattern') {
            return null
        }

        // Only show symbol suggestions if the query contains a context:, repo:
        // or file: filter.
        if (
            !tokens.some(token => {
                if (token.type !== 'filter') {
                    return false
                }
                return (
                    (isFilterOfType(token, FilterType.context) && token.value?.value !== 'global') ||
                    isFilterOfType(token, FilterType.repo) ||
                    isFilterOfType(token, FilterType.file)
                )
            })
        ) {
            return null
        }

        const includeType =
            !parsedQuery ||
            getRelevantTokens(
                parsedQuery,
                token.range,
                node => node.type === 'parameter' && resolveFilterMemoized(node.field)?.type === FilterType.type
            ).length === 0

        return cache.query(
            token.value,
            results => [
                {
                    title: 'Symbols',
                    options: limitUniqueOptions(results, DEFAULT_SUGGESTIONS_HIGH_PRI_LIST_SIZE, result =>
                        toSymbolSuggestion(result, includeType, token.range.start)
                    ),
                },
            ],
            parsedQuery,
            position
        )
    }
}

/**
 * A contextual cache not only uses the provided value to find suggestions but
 * also the current (parsed) query input.
 */
type ContextualCache<T, U> = CachedAsyncCompletionSource<T, U, [Node | null, number]>

interface Caches {
    repo: ContextualCache<Repo, FzfResultItem<Repo>>
    context: CachedAsyncCompletionSource<Context, FzfResultItem<Context>>
    file: ContextualCache<File, FzfResultItem<File>>
    symbol: ContextualCache<CodeSymbol, FzfResultItem<CodeSymbol>>
}

export interface SuggestionsSourceConfig {
    graphqlQuery: <T, V extends Record<string, any>>(query: string, variables: V) => Promise<T>
    authenticatedUser?: AuthenticatedUser | null
    isSourcegraphDotCom?: boolean
}

let sharedCaches: Caches | null = null

/**
 * Initializes and persists suggestion caches.
 */
function createCaches({ authenticatedUser, graphqlQuery }: SuggestionsSourceConfig): Caches {
    if (sharedCaches) {
        return sharedCaches
    }

    const cleanRegex = (value: string): string => value.replaceAll(/^\^|\\\.|\$$/g, '')

    const repoFzfOptions: FzfOptions<Repo> = {
        selector: item => item.name,
        tiebreakers: [starTiebraker],
        forward: false,
    }

    const contextFzfOptions: FzfOptions<Context> = {
        selector: item => item.spec,
        tiebreakers: [contextTiebraker],
    }

    const fileFzfOptions: FzfOptions<File> = {
        selector: item => item.path,
        forward: false,
        tiebreakers: [starTiebraker],
    }

    const symbolFzfOptions: FzfOptions<CodeSymbol> = {
        selector: item => item.name,
        tiebreakers: [byLengthAsc],
    }

    // Relevant query filters for file suggestions
    const fileFilters: Set<FilterType> = new Set([FilterType.repo, FilterType.rev, FilterType.context, FilterType.lang])
    const symbolFilters: Set<FilterType> = new Set([...fileFilters, FilterType.file])

    // TODO: Initialize outside to persist cache across page navigation
    return (sharedCaches = {
        repo: new CachedAsyncCompletionSource({
            // Repo queries are scoped to context: filters
            dataCacheKey: (parsedQuery, position) =>
                parsedQuery
                    ? buildSuggestionQuery(
                          parsedQuery,
                          { start: position, end: position },
                          node =>
                              isNonEmptyParameter(node) &&
                              resolveFilterMemoized(node.field)?.type === FilterType.context
                      )
                    : '',
            queryKey: (value, dataCacheKey = '') => `${dataCacheKey} type:repo count:50 repo:${value}`,
            async query(query) {
                const response = await graphqlQuery<SuggestionsRepoResult, SuggestionsRepoVariables>(REPOS_QUERY, {
                    query,
                })
                return response?.search?.results?.repositories.map(repository => [repository.name, repository]) ?? []
            },
            filter(repos, query) {
                const fzf = new Fzf(repos, repoFzfOptions)
                return fzf.find(cleanRegex(query))
            },
        }),

        context: new CachedAsyncCompletionSource({
            queryKey: value => `context:${value}`,
            async query(_key, value) {
                if (!authenticatedUser) {
                    return []
                }

                const response = await graphqlQuery<SuggestionsSearchContextResult, SuggestionsSearchContextVariables>(
                    SEARCH_CONTEXT_QUERY,
                    {
                        first: 20,
                        query: value,
                        namespaces: getUserSearchContextNamespaces(authenticatedUser),
                    }
                )
                return response.searchContexts.nodes.map(node => [
                    node.name,
                    {
                        name: node.name,
                        spec: node.spec,
                        default: node.viewerHasAsDefault,
                        starred: node.viewerHasStarred,
                        description: node.description,
                    },
                ])
            },
            filter(contexts, query) {
                const fzf = new Fzf(contexts, contextFzfOptions)
                const results = fzf.find(cleanRegex(query))
                if (query.trim() === '') {
                    // It seems we need to manually sort results if the query is
                    // empty to ensure that default and starred contexts are
                    // listed first.
                    results.sort(contextTiebraker)
                }
                return results
            },
        }),
        // File queries are scoped to context: and repo: filters
        file: new CachedAsyncCompletionSource({
            dataCacheKey: (parsedQuery, position) =>
                parsedQuery
                    ? buildSuggestionQuery(
                          parsedQuery,
                          { start: position, end: position },
                          node => isNonEmptyParameter(node) && containsFilterType(fileFilters, node.field)
                      )
                    : '',
            queryKey: (value, dataCacheKey = '') => `${dataCacheKey} type:file count:50 file:${value}`,
            async query(query) {
                const response = await graphqlQuery<SuggestionsFileResult, SuggestionsFileVariables>(FILE_QUERY, {
                    query,
                })
                return (
                    response.search?.results?.results?.reduce((results, result) => {
                        if (result.__typename === 'FileMatch') {
                            results.push([
                                result.file.path,
                                {
                                    path: result.file.path,
                                    repository: result.file.repository.name,
                                    stars: result.file.repository.stars,
                                    url: result.file.url,
                                },
                            ])
                        }
                        return results
                    }, [] as [string, File][]) ?? []
                )
            },
            filter(files, query) {
                const fzf = new Fzf(files, fileFzfOptions)
                return fzf.find(cleanRegex(query))
            },
        }),
        symbol: new CachedAsyncCompletionSource({
            dataCacheKey: (parsedQuery, position) =>
                parsedQuery
                    ? buildSuggestionQuery(
                          parsedQuery,
                          { start: position, end: position },
                          node => isNonEmptyParameter(node) && containsFilterType(symbolFilters, node.field)
                      )
                    : '',
            queryKey: (value, dataCacheKey = '') => `${dataCacheKey} type:symbol count:50 ${value}`,
            async query(query) {
                const response = await graphqlQuery<SuggestionsSymbolResult, SuggestionsSymbolVariables>(SYMBOL_QUERY, {
                    query,
                })
                return (
                    response.search?.results?.results?.reduce((results, result) => {
                        if (result.__typename === 'FileMatch') {
                            for (const symbol of result.symbols) {
                                results.push([
                                    symbol.url,
                                    {
                                        name: symbol.name,
                                        kind: symbol.kind,
                                        path: result.file.path,
                                        url: symbol.url,
                                    },
                                ])
                            }
                        }
                        return results
                    }, [] as [string, CodeSymbol][]) ?? []
                )
            },
            filter(files, query) {
                const fzf = new Fzf(files, symbolFzfOptions)
                return fzf.find(query)
            },
        }),
    })
}

/**
 * Main function of this module. It creates a suggestion source which internally
 * delegates to other sources.
 */
export const createSuggestionsSource = (config: SuggestionsSourceConfig): Source => {
    const caches = createCaches(config)

    const sources: InternalSource[] = [
        defaultSuggestions,
        filterValueSuggestions(caches),
        filterSuggestions,
        repoSuggestions(caches.repo),
        fileSuggestions(caches.file, config.isSourcegraphDotCom),
        symbolSuggestions(caches.symbol),
    ]

    return {
        query: (state, position) => {
            const queryInfo = getQueryInformation(state, position)

            function valid(state: EditorState, position: number): boolean {
                return queryInfo.token === getQueryInformation(state, position).token
            }

            const results = sources.map(source => source(queryInfo))
            const dummyResult = { result: [], valid }

            return combineResults([dummyResult, ...results])
        },
    }
}

function buildSuggestionQuery(query: Node, target: CharacterRange, filter: (node: Node) => boolean): string {
    return stringHuman(getRelevantTokens(query, target, filter))
}

function isNonEmptyParameter(node: Node): node is Parameter {
    return node.type === 'parameter' && !!node.value
}

function containsFilterType(filterTypes: Set<FilterType>, filterType: string): boolean {
    const resolvedFilter = resolveFilterMemoized(filterType)
    if (!resolvedFilter) {
        return false
    }
    return filterTypes.has(resolvedFilter.type)
}

function byStartDesc(itemA: FzfResultItem<unknown>, itemB: FzfResultItem<unknown>): number {
    return itemB.start - itemA.start
}

function shiftPositions(positions: Set<number>, amount: number): Set<number> {
    return new Set(Array.from(positions, position => position + amount))
}

function getFilterDescription(filter: ResolvedFilter, negated = false): string | undefined {
    if (!filter) {
        return undefined
    }
    return typeof filter.definition.description === 'function'
        ? filter.definition.description(negated)
        : filter.definition.description
}

/**
 * Returns a reduces list of unique options.
 */
function limitUniqueOptions<T>(values: T[], limit: number, mapper: (value: T) => Option): Option[] {
    const seen = new Set()
    const options: Option[] = []
    for (const value of values) {
        const option = mapper(value)
        if (!seen.has(option.label)) {
            seen.add(option.label)
            options.push(option)
        }

        if (options.length >= limit) {
            break
        }
    }
    return options
}
