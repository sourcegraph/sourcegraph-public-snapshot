import React from 'react'

import { EditorState } from '@codemirror/state'
import { mdiFilterOutline, mdiTextSearchVariant, mdiSourceRepository, mdiStar } from '@mdi/js'
import { extendedMatch, Fzf, FzfOptions, FzfResultItem } from 'fzf'
import { AuthenticatedUser } from 'src/auth'
import { SuggestionsRepoResult, SuggestionsRepoVariables } from 'src/graphql-operations'

import { tokenAt, tokens as queryTokens } from '@sourcegraph/branded'
// This module implements suggestions for the experimental search input
// eslint-disable-next-line no-restricted-imports
import {
    Group,
    Option,
    Target,
    Completion,
    Source,
    FilterOption,
    QueryOption,
    getEditorConfig,
    SuggestionResult,
} from '@sourcegraph/branded/src/search-ui/experimental'
import { gql } from '@sourcegraph/http-client'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { SearchContextProps } from '@sourcegraph/shared/src/search'
import { regexInsertText } from '@sourcegraph/shared/src/search/query/completion-utils'
import { FILTERS, FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { Filter, Token } from '@sourcegraph/shared/src/search/query/token'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'

/**
 * Used to organize the various sources that contribute to the final list of
 * suggestions.
 */
type InternalSource<T extends Token | undefined = Token | undefined> = (params: {
    token: T
    tokens: Token[]
    input: string
    position: number
}) => SuggestionResult | null

const none: any[] = []

const filterRenderer = (option: Option): React.ReactElement => React.createElement(FilterOption, { option })
const queryRenderer = (option: Option): React.ReactElement => React.createElement(QueryOption, { option })

function starTiebraker(a: { item: { stars: number } }, b: { item: { stars: number } }): number {
    return b.item.stars - a.item.stars
}

function contextTiebraker(a: { item: Context }, b: { item: Context }): number {
    return (b.item.starred || b.item.default ? 1 : 0) - (a.item.starred || a.item.default ? 1 : 0)
}

const REPOS_QUERY = gql`
    query SuggestionsRepo($query: String!) {
        search(patternType: regexp, query: $query) {
            results {
                repositories {
                    name
                    stars
                }
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

/**
 * Converts a Repo value to a (jump) target suggestion.
 */
function toRepoTarget({ item, positions }: FzfResultItem<Repo>): Target {
    return {
        type: 'target',
        icon: mdiSourceRepository,
        value: item.name,
        url: `/${item.name}`,
        matches: positions,
    }
}

/**
 * Converts a Repo value to a completion suggestion.
 */
function toRepoCompletion({ item, positions }: FzfResultItem<Repo>, from: number, to?: number): Completion {
    return {
        type: 'completion',
        icon: mdiSourceRepository,
        value: item.name,
        insertValue: regexInsertText(item.name, { globbing: false }) + ' ',
        matches: positions,
        from,
        to,
    }
}

/**
 * Converts a Context value to a completion suggestion.
 */
function toContextCompletion({ item, positions }: FzfResultItem<Context>, from: number, to?: number): Completion {
    let description = item.default ? 'Default' : ''
    if (item.description) {
        if (item.default) {
            description += '・'
        }
        description += item.description
    }

    return {
        type: 'completion',
        // Passing an empty string is a hack to draw an "empty" icon
        icon: item.starred ? mdiStar : ' ',
        value: item.spec,
        insertValue: item.spec + ' ',
        description,
        matches: positions,
        from,
        to,
    }
}

/**
 * Converts a filter to a completion suggestion.
 */
function toFilterCompletion(filter: FilterType, from: number, to?: number): Completion {
    const definition = FILTERS[filter]
    const description =
        typeof definition.description === 'function' ? definition.description(false) : definition.description
    return {
        type: 'completion',
        icon: mdiFilterOutline,
        render: filterRenderer,
        value: filter,
        insertValue: filter + ':',
        description,
        from,
        to,
    }
}

/**
 * If the query is not empty, this source will return a single command
 * suggestion which submits the query when selected.
 */
const currentQuery: InternalSource = ({ token, input }) => {
    if (token?.type === 'filter') {
        return null
    }

    let value = input
    let note = 'Search everywhere'

    const contextFilter = findFilter(input, FilterType.context, FilterKind.Global)

    if (contextFilter) {
        value = omitFilter(input, contextFilter)
        if (contextFilter.value?.value !== 'global') {
            note = `Search '${contextFilter.value?.value ?? ''}'`
        }
    }

    if (value.trim() === '') {
        return null
    }

    return {
        result: [
            {
                title: '',
                options: [
                    {
                        type: 'command',
                        icon: mdiTextSearchVariant,
                        value,
                        note,
                        apply: view => {
                            getEditorConfig(view.state).onSubmit()
                        },
                        render: queryRenderer,
                    },
                ],
            },
        ],
    }
}

const FILTER_SUGGESTIONS = new Fzf(Object.keys(FILTERS) as FilterType[], { match: extendedMatch })
const DEFAULT_FILTERS: FilterType[] = [FilterType.repo, FilterType.context, FilterType.lang, FilterType.type]
const RELATED_FILTERS: Partial<Record<FilterType, (filter: Filter) => FilterType[]>> = {
    [FilterType.type]: filter => {
        switch (filter.value?.value) {
            case 'diff':
            case 'commit':
                return [FilterType.author, FilterType.before, FilterType.after, FilterType.message]
        }
        return []
    },
}

/**
 * Returns filter completion suggestions for the current term at the cursor. If
 * there is no term a small list of suggested filters is returned.
 */
const filterSuggestions: InternalSource = ({ tokens, token, position }) => {
    let options: Group['options'] = []

    if (!token || token.type === 'whitespace') {
        const filters = DEFAULT_FILTERS
            // Add related filters
            .concat(
                tokens.flatMap(token =>
                    token.type === 'filter' ? RELATED_FILTERS[token.field.value as FilterType]?.(token) ?? none : none
                )
            )
            // Remove existing filters
            .filter(filter => !tokens.some(token => token.type === 'filter' && token.field.value === filter))

        options = filters.map(filter => toFilterCompletion(filter, position))
    } else if (token?.type === 'pattern') {
        // ^ triggers a prefix match
        options = FILTER_SUGGESTIONS.find('^' + token.value).map(entry => ({
            ...toFilterCompletion(entry.item, token.range.start, token.range.end),
            matches: entry.positions,
        }))
    }

    return options.length > 0 ? { result: [{ title: 'Narrow your search', options }] } : null
}

/**
 * Returns static and dynamic completion suggestions for filters when completing
 * a filter value.
 */
function filterValueSuggestions(caches: Caches): InternalSource {
    return ({ token }) => {
        if (token?.type !== 'filter') {
            return null
        }
        const resolvedFilter = resolveFilter(token.field.value)
        const value = token.value?.value ?? ''
        const from = token.value?.range.start ?? token.range.end
        const to = token.value?.range.end

        switch (resolvedFilter?.definition.suggestions) {
            case 'repo': {
                return caches.repo.query(value, entries => [
                    {
                        title: 'Repositories',
                        options: entries.slice(0, 25).map(item => toRepoCompletion(item, from, to)),
                    },
                ])
            }

            default: {
                switch (resolvedFilter?.type) {
                    // Some filters are not defined to have dynamic suggestions,
                    // we need to handle these here explicitly. We can't change
                    // the filter definition without breaking the current
                    // search input.
                    case FilterType.context:
                        return caches.context.query(value, entries => {
                            entries = value.trim() === '' ? entries.slice(0, 10) : entries
                            return [
                                {
                                    title: 'Search contexts',
                                    options: entries.map(entry => toContextCompletion(entry, from, to)),
                                },
                                {
                                    title: 'Actions',
                                    options: [
                                        {
                                            type: 'target',
                                            value: 'Manage contexts',
                                            description: 'Add, edit, remove search contexts',
                                            note: 'Got to /contexts',
                                            url: '/contexts',
                                        },
                                    ],
                                },
                            ]
                        })
                    default: {
                        const suggestions = staticFilterValueSuggestions(token)
                        return suggestions ? { result: [suggestions] } : null
                    }
                }
            }
        }
    }
}

function staticFilterValueSuggestions(token?: Token): Group | null {
    if (token?.type !== 'filter') {
        return null
    }

    const resolvedFilter = resolveFilter(token.field.value)
    if (!resolvedFilter?.definition.discreteValues) {
        return null
    }

    const value = token.value
    let options: Completion[] = resolvedFilter.definition.discreteValues(token.value, false).map(value => ({
        type: 'completion',
        from: token.value?.range.start ?? token.range.end,
        to: token.value?.range.end,
        value: value.label,
        insertValue: (value.insertText ?? value.label) + ' ',
    }))

    if (value && value.value !== '') {
        const fzf = new Fzf(options, { selector: option => option.value })
        options = fzf.find(value.value).map(match => ({ ...match.item, matches: match.positions }))
    }

    // TODO: Determine appropriate title
    return options.length > 0 ? { title: '', options } : null
}

/**
 * Returns repository (jump) target suggestions matching the term at the cursor,
 * but only if the query doens't already contain a 'repo:' filter.
 */
function repoSuggestions(cache: Cache<Repo, FzfResultItem<Repo>>): InternalSource {
    return ({ token, tokens }) => {
        const showRepoSuggestions =
            token?.type === 'pattern' && !tokens.some(token => token.type === 'filter' && token.field.value === 'repo')
        if (!showRepoSuggestions) {
            return null
        }

        return cache.query(token.value, results => [
            {
                title: 'Repositories',
                options: results.slice(0, 5).map(toRepoTarget),
            },
        ])
    }
}

interface Caches {
    repo: Cache<Repo, FzfResultItem<Repo>>
    context: Cache<Context, FzfResultItem<Context>>
}

interface SuggestionsSourceConfig
    extends Pick<SearchContextProps, 'fetchSearchContexts' | 'getUserSearchContextNamespaces'> {
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
    authenticatedUser?: AuthenticatedUser | null
}

/**
 * Main function of this module. It creates a suggestion source which internally
 * delegates to other sources.
 */
export const createSuggestionsSource = ({
    platformContext,
    authenticatedUser,
    fetchSearchContexts,
    getUserSearchContextNamespaces,
}: SuggestionsSourceConfig): Source => {
    const cleanRegex = (value: string): string => value.replace(/^\^|\\\.|\$$/g, '')

    const repoFzfOptions: FzfOptions<Repo> = {
        selector: item => item.name,
        tiebreakers: [starTiebraker],
    }

    const contextFzfOptions: FzfOptions<Context> = {
        selector: item => item.spec,
        tiebreakers: [contextTiebraker],
    }

    // TODO: Initialize outside to persist cache across page navigation
    const caches: Caches = {
        repo: new Cache({
            queryKey: value => `type:repo count:50 repo:${value}`,
            async query(query) {
                const response = await platformContext
                    .requestGraphQL<SuggestionsRepoResult, SuggestionsRepoVariables>({
                        request: REPOS_QUERY,
                        variables: { query },
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                return (
                    response.data?.search?.results?.repositories.map(repository => [repository.name, repository]) || []
                )
            },
            filter(repos, query) {
                const fzf = new Fzf(repos, repoFzfOptions)
                return fzf.find(cleanRegex(query))
            },
        }),

        context: new Cache({
            queryKey: value => `context:${value}`,
            async query(_key, value) {
                if (!authenticatedUser) {
                    return []
                }

                const response = await fetchSearchContexts({
                    first: 50,
                    query: value,
                    platformContext,
                    namespaces: getUserSearchContextNamespaces(authenticatedUser),
                }).toPromise()
                return response.nodes.map(node => [
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
                    // We need to manually sort results if the query is empty to
                    // ensure that default and starred contexts are listed
                    // first.
                    results.sort(contextTiebraker)
                }
                return results
            },
        }),
    }

    const sources: InternalSource[] = [
        currentQuery,
        filterValueSuggestions(caches),
        filterSuggestions,
        repoSuggestions(caches.repo),
    ]

    return (state, position) => {
        const tokens = collapseOpenFilterValues(queryTokens(state), state.sliceDoc())
        const token = tokenAt(tokens, position)
        const input = state.sliceDoc()

        function valid(state: EditorState, position: number): boolean {
            const tokens = collapseOpenFilterValues(queryTokens(state), state.sliceDoc())
            return token === tokenAt(tokens, position)
        }

        const params = { token, tokens, input, position }
        const results = sources.map(source => source(params))
        const dummyResult = { result: [], valid }

        return combineResults([dummyResult, ...results])
    }
}

interface CacheConfig<T, U> {
    queryKey(value: string): string
    query(key: string, value: string): Promise<[string, T][]>
    filter(entries: T[], value: string): U[]
}

/**
 * This class handles creating suggestion results that include cached values (if
 * available) and updates the cache with new results from new queries.
 */
class Cache<T, U> {
    private queryCache = new Map<string, Promise<void>>()
    private dataCache = new Map<string, T>()

    constructor(private config: CacheConfig<T, U>) {}

    public query(value: string, mapper: (values: U[]) => Group[]): ReturnType<InternalSource> {
        const next: SuggestionResult['next'] = () => {
            const key = this.config.queryKey(value)
            let result = this.queryCache.get(key)

            if (!result) {
                result = this.config.query(key, value).then(entries => {
                    for (const [key, entry] of entries) {
                        if (!this.dataCache.has(key)) {
                            this.dataCache.set(key, entry)
                        }
                    }
                })
            }

            if (!this.queryCache.has(key)) {
                this.queryCache.set(key, result)
            }

            return result.then(() => ({ result: mapper(this.cachedData(value)) }))
        }

        return {
            result: mapper(this.cachedData(value)),
            next,
        }
    }

    private cachedData(value: string): U[] {
        return this.config.filter(Array.from(this.dataCache.values()), value)
    }
}

// Helper function to convert filter values that start with a quote but are not
// closed yet (e.g. author:"firstname lastna|) to a single filter token to
// prevent irrelevant suggestions.
function collapseOpenFilterValues(tokens: Token[], input: string): Token[] {
    const result: Token[] = []
    let openFilter: Filter | null = null
    let hold: Token[] = []

    function mergeFilter(filter: Filter, values: Token[]): Filter {
        if (!filter.value?.value) {
            // For simplicity but this should never occure
            return filter
        }
        const end = values[values.length - 1]?.range.end ?? filter.value.range.end
        return {
            ...filter,
            range: {
                start: filter.range.start,
                end,
            },
            value: {
                ...filter.value,
                range: {
                    start: filter.value.range.start,
                    end,
                },
                value:
                    filter.value.value + values.map(token => input.slice(token.range.start, token.range.end)).join(''),
            },
        }
    }

    for (const token of tokens) {
        switch (token.type) {
            case 'filter':
                {
                    if (token.value?.value.startsWith('"') && !token.value.quoted) {
                        openFilter = token
                    } else {
                        if (openFilter?.value) {
                            result.push(mergeFilter(openFilter, hold))
                            openFilter = null
                            hold = []
                        }
                        result.push(token)
                    }
                }
                break
            case 'pattern':
            case 'whitespace':
                if (openFilter) {
                    hold.push(token)
                } else {
                    result.push(token)
                }
                break
            default:
                if (openFilter?.value) {
                    result.push(mergeFilter(openFilter, hold))
                    openFilter = null
                    hold = []
                }
                result.push(token)
        }
    }

    if (openFilter?.value) {
        result.push(mergeFilter(openFilter, hold))
    }

    return result
}

/**
 * Takes multiple suggestion results and combines the groups of each of them.
 * The order of items within a group is determined by the order of results.
 */
function combineResults(results: (SuggestionResult | null)[]): SuggestionResult {
    const options: Record<Group['title'], Group['options'][]> = {}
    let hasValid = false
    let hasNext = false

    for (const result of results) {
        if (!result) {
            continue
        }
        for (const group of result.result) {
            if (!options[group.title]) {
                options[group.title] = []
            }
            options[group.title].push(group.options)
        }
        if (result.next) {
            hasNext = true
        }
        if (result.valid) {
            hasValid = true
        }
    }

    const staticResult: SuggestionResult = {
        result: Object.entries(options).map(([title, options]) => ({ title, options: options.flat() })),
    }

    if (hasValid) {
        staticResult.valid = (...args) => results.every(result => result?.valid?.(...args) ?? false)
    }
    if (hasNext) {
        staticResult.next = () => Promise.all(results.map(result => result?.next?.() ?? result)).then(combineResults)
    }

    return staticResult
}
