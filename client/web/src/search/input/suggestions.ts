import { EditorState } from '@codemirror/state'
import { mdiFilterOutline, mdiTextSearchVariant, mdiSourceRepository, mdiStar, mdiFileOutline } from '@mdi/js'
import { byLengthAsc, extendedMatch, Fzf, FzfOptions, FzfResultItem } from 'fzf'

import { tokenAt, tokens as queryTokens } from '@sourcegraph/branded'
// This module implements suggestions for the experimental search input
// eslint-disable-next-line no-restricted-imports
import {
    Group,
    Option,
    Source,
    getEditorConfig,
    SuggestionResult,
    submitQueryInfo,
    queryRenderer,
    filterRenderer,
    filterValueRenderer,
    shortenPath,
} from '@sourcegraph/branded/src/search-ui/experimental'
import { getParsedQuery } from '@sourcegraph/branded/src/search-ui/input/codemirror/parsedQuery'
import { isDefined } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { SearchContextProps } from '@sourcegraph/shared/src/search'
import { regexInsertText } from '@sourcegraph/shared/src/search/query/completion-utils'
import { FILTERS, FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { Node, OperatorKind } from '@sourcegraph/shared/src/search/query/parser'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { CharacterRange, Filter, PatternKind, Token } from '@sourcegraph/shared/src/search/query/token'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { getSymbolIconSVGPath } from '@sourcegraph/shared/src/symbols/symbolIcons'

import { AuthenticatedUser } from '../../auth'
import {
    SuggestionsRepoResult,
    SuggestionsRepoVariables,
    SuggestionsFileResult,
    SuggestionsFileVariables,
    SuggestionsSymbolResult,
    SuggestionsSymbolVariables,
    SymbolKind,
} from '../../graphql-operations'

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

const SYMBOL_QUERY = gql`
    query SuggestionsSymbol($query: String!) {
        search(patternType: regexp, query: $query) {
            results {
                results {
                    ... on FileMatch {
                        __typename
                        file {
                            path
                        }
                        symbols {
                            kind
                            url
                            name
                        }
                    }
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
    option.action.name = 'Add'
    option.alternativeAction = {
        type: 'goto',
        url: `/${result.item.name}`,
    }
    option.render = filterValueRenderer
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
    return {
        label: valuePrefix + item.name,
        matches: positions,
        icon: mdiSourceRepository,
        action: {
            type: 'completion',
            insertValue: valuePrefix + regexInsertText(item.name, { globbing: false }) + ' ',
            from,
            to,
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
function toFilterCompletion(filter: FilterType, from: number, to?: number): Option {
    const definition = FILTERS[filter]
    const description =
        typeof definition.description === 'function' ? definition.description(false) : definition.description
    return {
        label: filter,
        icon: mdiFilterOutline,
        render: filterRenderer,
        description,
        action: {
            type: 'completion',
            insertValue: filter + ':',
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
    return {
        label: valuePrefix + item.path,
        icon: mdiFileOutline,
        description: item.repository,
        matches: positions,
        action: {
            type: 'completion',
            insertValue: valuePrefix + regexInsertText(item.path, { globbing: false }) + ' ',
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
    option.action.name = 'Add'
    option.alternativeAction = {
        type: 'goto',
        url: result.item.url,
    }
    option.render = filterValueRenderer
    return option
}

/**
 * Converts a File value to a (jump) target suggestion.
 */
function toSymbolSuggestion({ item, positions }: FzfResultItem<CodeSymbol>, from: number, to?: number): Option {
    return {
        label: item.name,
        matches: positions,
        description: shortenPath(item.path, 20),
        icon: getSymbolIconSVGPath(item.kind),
        action: {
            type: 'completion',
            insertValue: item.name + ' type:symbol ',
            from,
            to,
        },
        alternativeAction: {
            type: 'goto',
            url: item.url,
        },
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

    let label = input
    let actionName = 'Search everywhere'

    const contextFilter = findFilter(input, FilterType.context, FilterKind.Global)

    if (contextFilter) {
        label = omitFilter(input, contextFilter)
        if (contextFilter.value?.value !== 'global') {
            actionName = `Search '${contextFilter.value?.value ?? ''}'`
        }
    }

    if (label.trim() === '') {
        return null
    }

    return {
        result: [
            {
                title: '',
                options: [
                    {
                        icon: mdiTextSearchVariant,
                        label,
                        action: {
                            type: 'command',
                            name: actionName,
                            apply: view => {
                                getEditorConfig(view.state).onSubmit()
                            },
                        },
                        render: queryRenderer,
                        info: submitQueryInfo,
                    },
                ],
            },
        ],
    }
}

const FILTER_SUGGESTIONS = new Fzf(Object.keys(FILTERS) as FilterType[], { match: extendedMatch })
// These are the filters shown when the query input is empty or the cursor is at
// at whitespace token.
const DEFAULT_FILTERS: FilterType[] = [FilterType.repo, FilterType.context, FilterType.lang, FilterType.type]
// If the query contains one of the listed filters, suggest these filters
// too.
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
 * there is no term a small list of default filters is returned. Filters are
 * matched by prefix.
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

const contextActions: Group = {
    title: 'Actions',
    options: [
        {
            label: 'Manage contexts',
            description: 'Add, edit, remove search contexts',
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
        const resolvedFilter = resolveFilter(token.field.value)
        const value = token.value?.value ?? ''
        const from = token.value?.range.start ?? token.range.end
        const to = token.value?.range.end

        switch (resolvedFilter?.definition.suggestions) {
            case 'repo': {
                return caches.repo.query(
                    value,
                    entries => [
                        {
                            title: 'Repositories',
                            options: entries.slice(0, 25).map(item => toRepoCompletion(item, from, to)),
                        },
                    ],
                    parsedQuery,
                    position
                )
            }

            case 'path': {
                return caches.file.query(
                    value,
                    entries => [
                        {
                            title: 'Files',
                            options: entries.map(item => toFileCompletion(item, from, to)).slice(0, 25),
                        },
                    ],
                    parsedQuery,
                    position
                )
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
                                contextActions,
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
    let options: Option[] = resolvedFilter.definition.discreteValues(token.value, false).map(value => ({
        label: value.label,
        action: {
            type: 'completion',
            from: token.value?.range.start ?? token.range.end,
            to: token.value?.range.end,
            insertValue: (value.insertText ?? value.label) + ' ',
        },
    }))

    if (value && value.value !== '') {
        const fzf = new Fzf(options, { selector: option => option.label })
        options = fzf.find(value.value).map(match => ({ ...match.item, matches: match.positions }))
    }

    // TODO: Determine appropriate title
    return options.length > 0 ? { title: '', options } : null
}

/**
 * Returns repository (jump) target suggestions matching the term at the cursor,
 * but only if the query doesn't already contain a 'repo:' filter.
 */
function repoSuggestions(cache: Caches['repo']): InternalSource {
    return ({ token, tokens, parsedQuery, position }) => {
        const showRepoSuggestions =
            token?.type === 'pattern' && !tokens.some(token => token.type === 'filter' && token.field.value === 'repo')
        if (!showRepoSuggestions) {
            return null
        }

        return cache.query(
            token.value,
            results => [
                {
                    title: 'Repositories',
                    options: results.slice(0, 3).map(result => toRepoSuggestion(result, token.range.start)),
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
                tokens.some(
                    token =>
                        token.type === 'filter' &&
                        ((token.field.value === 'context' && token.value?.value !== 'global') ||
                            token.field.value === 'repo')
                ))

        if (!showFileSuggestions) {
            return null
        }

        return cache.query(
            token.value,
            results => [
                {
                    title: 'Files',
                    options: results.slice(0, 5).map(result => toFileSuggestion(result, token.range.start)),
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
function symbolSuggestions(cache: Caches['symbol'], isSourcegraphDotCom?: boolean): InternalSource {
    return ({ token, tokens, parsedQuery, position }) => {
        if (token?.type !== 'pattern') {
            return null
        }

        // Only show symbol suggestions if the query contains a context:, repo:
        // or file: filter. On dotcom the context must by different from
        // "global".

        if (
            !tokens.some(
                token =>
                    token.type === 'filter' &&
                    ((token.field.value === 'context' && (!isSourcegraphDotCom || token.value?.value !== 'global')) ||
                        token.field.value === 'repo' ||
                        token.field.value === 'file')
            )
        ) {
            return null
        }

        return cache.query(
            token.value,
            results => [
                {
                    title: 'Symbols',
                    options: results.slice(0, 5).map(result => toSymbolSuggestion(result, token.range.start)),
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
type ContextualCache<T, U> = Cache<T, U, [Node | null, number]>

interface Caches {
    repo: ContextualCache<Repo, FzfResultItem<Repo>>
    context: Cache<Context, FzfResultItem<Context>>
    file: ContextualCache<File, FzfResultItem<File>>
    symbol: ContextualCache<CodeSymbol, FzfResultItem<CodeSymbol>>
}

interface SuggestionsSourceConfig
    extends Pick<SearchContextProps, 'fetchSearchContexts' | 'getUserSearchContextNamespaces'> {
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
    authenticatedUser?: AuthenticatedUser | null
    isSourcegraphDotCom?: boolean
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
    isSourcegraphDotCom,
}: SuggestionsSourceConfig): Source => {
    const cleanRegex = (value: string): string => value.replace(/^\^|\\\.|\$$/g, '')

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
    const caches: Caches = {
        repo: new Cache({
            // Repo queries are scoped to context: filters
            dataCacheKey: (parsedQuery, position) =>
                parsedQuery
                    ? buildSuggestionQuery(
                          parsedQuery,
                          { start: position, end: position },
                          token => token.type === 'parameter' && !!token.value && token.field === 'context'
                      )
                    : '',
            queryKey: (value, dataCacheKey = '') => `${dataCacheKey} type:repo count:50 repo:${value}`,
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
                    // It seems we need to manually sort results if the query is
                    // empty to ensure that default and starred contexts are
                    // listed first.
                    results.sort(contextTiebraker)
                }
                return results
            },
        }),
        // File queries are scoped to context: and repo: filters
        file: new Cache({
            dataCacheKey: (parsedQuery, position) =>
                parsedQuery
                    ? buildSuggestionQuery(
                          parsedQuery,
                          { start: position, end: position },
                          token =>
                              token.type === 'parameter' && !!token.value && fileFilters.has(token.field as FilterType)
                      )
                    : '',
            queryKey: (value, dataCacheKey = '') => `${dataCacheKey} type:file count:50 file:${value}`,
            async query(query) {
                const response = await platformContext
                    .requestGraphQL<SuggestionsFileResult, SuggestionsFileVariables>({
                        request: FILE_QUERY,
                        variables: { query },
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                return (
                    response.data?.search?.results?.results?.reduce((results, result) => {
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
        symbol: new Cache({
            dataCacheKey: (parsedQuery, position) =>
                parsedQuery
                    ? buildSuggestionQuery(
                          parsedQuery,
                          { start: position, end: position },
                          token =>
                              token.type === 'parameter' &&
                              !!token.value &&
                              symbolFilters.has(token.field as FilterType)
                      )
                    : '',
            queryKey: (value, dataCacheKey = '') => `${dataCacheKey} type:symbol count:50 ${value}`,
            async query(query) {
                const response = await platformContext
                    .requestGraphQL<SuggestionsSymbolResult, SuggestionsSymbolVariables>({
                        request: SYMBOL_QUERY,
                        variables: { query },
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                return (
                    response.data?.search?.results?.results?.reduce((results, result) => {
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
    }

    const sources: InternalSource[] = [
        currentQuery,
        filterValueSuggestions(caches),
        filterSuggestions,
        repoSuggestions(caches.repo),
        fileSuggestions(caches.file, isSourcegraphDotCom),
        symbolSuggestions(caches.symbol, isSourcegraphDotCom),
    ]

    return (state, position) => {
        const parsedQuery = getParsedQuery(state)
        const tokens = collapseOpenFilterValues(queryTokens(state), state.sliceDoc())
        const token = tokenAt(tokens, position)
        const input = state.sliceDoc()

        function valid(state: EditorState, position: number): boolean {
            const tokens = collapseOpenFilterValues(queryTokens(state), state.sliceDoc())
            return token === tokenAt(tokens, position)
        }

        const params = { token, tokens, input, position, parsedQuery }
        const results = sources.map(source => source(params))
        const dummyResult = { result: [], valid }

        return combineResults([dummyResult, ...results])
    }
}

interface CacheConfig<T, U, E extends any[] = []> {
    /**
     * Returns a string that uniquely identifies this query (which is often just
     * the query itself). If the same request is made again the existing result
     * is reused.
     */
    queryKey(value: string, dataCacheKey?: string): string
    /**
     * Fetch data. queryKey is the value return by the queryKey function and
     * value is the term that's currently completed. Returns a list of [key,
     * value] tuples. The key of these tuples is used to uniquly identify a
     * value the data cache.
     */
    query(queryKey: string, value: string): Promise<[string, T][]>
    /**
     * This function filters and ranks all cache values (entries) by value.
     */
    filter(entries: T[], value: string): U[]
    /**
     * If provided data values are bucketed into different "cache groups", keyed
     * by the return value of this function.
     */
    dataCacheKey?(...extraArgs: E): string
}

/**
 * This class handles creating suggestion results that include cached values (if
 * available) and updates the cache with new results from new queries.
 */
class Cache<T, U, E extends any[] = []> {
    private queryCache = new Map<string, Promise<void>>()
    private dataCache = new Map<string, T>()
    private dataCacheByQuery = new Map<string, Map<string, T>>()

    constructor(private config: CacheConfig<T, U, E>) {}

    public query(value: string, mapper: (values: U[]) => Group[], ...extraArgs: E): ReturnType<InternalSource> {
        // The dataCacheKey could possibly just be an argument to query. However
        // that would require callsites to remember to pass the value. Doing it
        // this way we get a bit more type safety.
        const dataCacheKey = this.config.dataCacheKey?.(...extraArgs)
        const queryKey = this.config.queryKey(value, dataCacheKey)
        let dataCache = this.dataCache
        if (dataCacheKey) {
            dataCache = this.dataCacheByQuery.get(dataCacheKey) ?? new Map<string, T>()
            if (!this.dataCacheByQuery.has(dataCacheKey)) {
                this.dataCacheByQuery.set(dataCacheKey, dataCache)
            }
        }
        return {
            result: mapper(this.cachedData(value, dataCache)),
            next: () => {
                let result = this.queryCache.get(queryKey)

                if (!result) {
                    result = this.config.query(queryKey, value).then(entries => {
                        for (const [key, entry] of entries) {
                            if (!dataCache.has(key)) {
                                dataCache.set(key, entry)
                            }
                        }
                    })

                    this.queryCache.set(queryKey, result)
                }

                return result.then(() => ({ result: mapper(this.cachedData(value, dataCache)) }))
            },
        }
    }

    private cachedData(value: string, cache = this.dataCache): U[] {
        return this.config.filter(Array.from(cache.values()), value)
    }
}

const placeholderRange: CharacterRange = { start: 0, end: 0 }

/**
 * This function processes a given query in a top-down manner and removes any
 * patterns and filters that cannot affect the token at the target character
 * range.
 * This is relatively straighforward: We only keep tokens that represent
 * whitelisted filters and which are direct children of an AND branch.
 * Everything else is discarded.
 */
function buildSuggestionQuery(query: Node, target: CharacterRange, filter: (node: Node) => boolean): string {
    function processNode(node: Node): Node | null {
        switch (node.type) {
            case 'parameter':
            case 'pattern':
                return filter(node) ? node : null
            case 'sequence': {
                const nodes = node.nodes.map(processNode).filter(isDefined)
                return nodes.length > 0 ? { type: 'sequence', nodes, range: placeholderRange } : null
            }
            case 'operator': {
                switch (node.kind) {
                    case OperatorKind.Or: {
                        // If one operand contains the target branche we only
                        // need to keep that operand (the other branch is
                        // irrelevant). But if no operand contains the target
                        // range we need to process all nodes and assume that
                        // this token is ANDed at some level with the target
                        // range.
                        //
                        // Examples:
                        //
                        // filter:a filter:b OR filter:|
                        // ^^^^^^^^^^^^^^^^^
                        //      discard
                        //
                        // (filter:a or filter:b) filter:|
                        // ^^^^^^^^^^^^^^^^^^^^^^
                        // needs to be preserved
                        const operand = node.operands.find(
                            node => node.range.start <= target.start && node.range.end >= target.end
                        )

                        if (operand) {
                            return processNode(operand)
                        }
                        // NOTE: Intentional fallthrough since the logic is the
                        // same.
                    }
                    case OperatorKind.And: {
                        const operands = node.operands.map(processNode).filter(isDefined)
                        switch (operands.length) {
                            case 0:
                                return null
                            case 1:
                                return operands[0]
                            default:
                                return {
                                    type: 'operator',
                                    // needs to be node.kind to properly handle
                                    // fallthrough case.
                                    kind: node.kind,
                                    operands,
                                    range: placeholderRange,
                                }
                        }
                    }
                    case OperatorKind.Not: {
                        if (node.operands.length === 0) {
                            return null
                        }
                        const operand = processNode(node.operands[0])
                        if (!operand) {
                            return null
                        }
                        return { type: 'operator', kind: node.kind, operands: [operand], range: placeholderRange }
                    }
                }
            }
        }
    }

    const result = processNode(query)
    return result ? printParsedQuery(result).join('') : ''
}

function printParsedQuery(node: Node, buffer: string[] = []): string[] {
    switch (node.type) {
        case 'pattern':
            // TODO: quoted, negated, ...
            switch (node.kind) {
                case PatternKind.Regexp:
                    buffer.push('/', node.value, '/')
                    return buffer
                default:
                    buffer.push(node.value)
                    return buffer
            }
        case 'parameter': {
            if (node.negated) {
                buffer.push('-')
            }
            buffer.push(node.field, ':', node.value)
            return buffer
        }
        case 'sequence': {
            for (const operand of node.nodes) {
                printParsedQuery(operand, buffer)
                buffer.push(' ')
            }
            return buffer
        }
        case 'operator': {
            buffer.push(
                ' (',
                node.operands.map(operand => printParsedQuery(operand).join('')).join(` ${node.kind} `),
                ') '
            )
            return buffer
        }
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
