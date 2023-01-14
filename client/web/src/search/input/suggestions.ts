import React from 'react'

import { EditorState } from '@codemirror/state'
import { mdiFilterOutline, mdiTextSearchVariant, mdiSourceRepository } from '@mdi/js'
import { extendedMatch, Fzf, FzfOptions, FzfResultItem } from 'fzf'
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
} from '@sourcegraph/branded/src/search-ui/experimental'
import { getDocumentNode, gql } from '@sourcegraph/http-client'
import { regexInsertText } from '@sourcegraph/shared/src/search/query/completion-utils'
import { FILTERS, FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { Filter, Token } from '@sourcegraph/shared/src/search/query/token'

import { getWebGraphQLClient } from '../../backend/graphql'

const filterRenderer = (option: Option): React.ReactElement => React.createElement(FilterOption, { option })
const queryRenderer = (option: Option): React.ReactElement => React.createElement(QueryOption, { option })
const none: any[] = []

const FILTER_SUGGESTIONS = new Fzf(Object.keys(FILTERS) as FilterType[], { match: extendedMatch })
const DEFAULT_FILTERS: FilterType[] = [FilterType.repo, FilterType.lang, FilterType.type]
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

export const filterSuggestions = (tokens: Token[], token: Token | undefined, position: number): Option[] => {
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

        return filters.map(filter => toFilterCompletion(filter, position))
    }

    if (token?.type === 'pattern') {
        // ^ triggers a prefix match
        return FILTER_SUGGESTIONS.find('^' + token.value).map(entry => ({
            ...toFilterCompletion(entry.item, token.range.start, token.range.end),
            matches: entry.positions,
        }))
    }
    return []
}

export const staticFilterValueSuggestions = (token?: Token): Group | null => {
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

function starTiebraker(a: { item: { stars: number } }, b: { item: { stars: number } }): number {
    return b.item.stars - a.item.stars
}

const repoFzfOptions: FzfOptions<Repo> = {
    selector: item => item.name,
    tiebreakers: [starTiebraker],
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
const repoCache: Map<string, Repo> = new Map()
const queryCache: Map<string, Promise<Repo[]>> = new Map()

function cachedRepos<T>(value: string, mapper: (item: FzfResultItem<Repo>) => T): T[] {
    const fzf = new Fzf([...repoCache.values()], repoFzfOptions)
    return fzf.find(value).map(mapper)
}

async function dynamicRepos<T>(value: string, mapper: (item: FzfResultItem<Repo>) => T): Promise<T[]> {
    const query = `type:repo count:50 repo:${value}`
    const repositories =
        queryCache.get(query) ??
        getWebGraphQLClient()
            .then(client =>
                client.query<SuggestionsRepoResult, SuggestionsRepoVariables>({
                    query: getDocumentNode(REPOS_QUERY),
                    variables: { query },
                })
            )
            .then(response =>
                (response.data?.search?.results?.repositories || []).map(({ name, stars }) => {
                    const repo = { name, stars }
                    if (!repoCache.has(name)) {
                        repoCache.set(name, repo)
                    }
                    return repo
                })
            )

    if (!queryCache.has(query)) {
        queryCache.set(query, repositories)
    }

    await repositories
    // Remove common regex special characters
    const cleanValue = value.replace(/^\^|\\\.|\$$/g, '')
    return cachedRepos(cleanValue, mapper)
}

function toRepoTarget(item: FzfResultItem<Repo>): Target {
    return {
        type: 'target',
        icon: mdiSourceRepository,
        value: item.item.name,
        url: `/${item.item.name}`,
        matches: item.positions,
    }
}

function toRepoCompletion(item: FzfResultItem<Repo>, from: number, to?: number): Completion {
    return {
        type: 'completion',
        icon: mdiSourceRepository,
        value: item.item.name,
        insertValue: regexInsertText(item.item.name, { globbing: false }) + ' ',
        matches: item.positions,
        from,
        to,
    }
}

export const dynamicRepoSuggestions = async (token?: Token): Promise<Target[]> => {
    if (token?.type !== 'pattern') {
        return []
    }

    return dynamicRepos(token.value, toRepoTarget)
}

const cachedRepoSuggestions = (token?: Token): Target[] => {
    if (token?.type !== 'pattern') {
        return []
    }

    return cachedRepos(token.value, toRepoTarget)
}

/**
 * Returns dynamic suggestions for filter values.
 */
function filterValueSuggestions(token: Token | undefined): ReturnType<Source> | null {
    if (token?.type === 'filter') {
        const resolvedFilter = resolveFilter(token.field.value)
        const value = token.value?.value ?? ''
        const from = token.value?.range.start ?? token.range.end
        const to = token.value?.range.end

        function valid(state: EditorState, position: number): boolean {
            const tokens = collapseOpenFilterValues(queryTokens(state), state.sliceDoc())
            return token === tokenAt(tokens, position)
        }

        switch (resolvedFilter?.definition.suggestions) {
            case 'repo': {
                const repos: Option[] = cachedRepos(value, item => toRepoCompletion(item, from, to)).slice(0, 25)
                return {
                    valid,
                    result: [{ title: 'Repositories', options: repos }],
                    next: () =>
                        dynamicRepos(value, item => toRepoCompletion(item, from, to)).then(entries => ({
                            valid,
                            result: [{ title: 'Repositories', options: entries.slice(0, 25) }],
                        })),
                }
            }

            default: {
                const suggestions = staticFilterValueSuggestions(token)
                return suggestions ? { result: [suggestions] } : null
            }
        }
    }
    return null
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

export const suggestions: Source = (state, pos) => {
    const tokens = collapseOpenFilterValues(queryTokens(state), state.sliceDoc())
    const token = tokenAt(tokens, pos)

    function valid(state: EditorState, position: number): boolean {
        const tokens = collapseOpenFilterValues(queryTokens(state), state.sliceDoc())
        return token === tokenAt(tokens, position)
    }

    const result: Group[] = []

    // Only show filter value completions if they are available
    const filterValueResult = filterValueSuggestions(token)
    if (filterValueResult) {
        return filterValueResult
    }

    // Default options
    const showRepoSuggestions = !tokens.some(token => token.type === 'filter' && token.field.value === 'repo')

    // Add running the current query as first/initial command
    const currentQuery = state.sliceDoc()
    if (currentQuery.trim() !== '') {
        result.push({
            title: '',
            options: [
                {
                    type: 'command',
                    icon: mdiTextSearchVariant,
                    value: state.sliceDoc(),
                    note: 'Search everywhere',
                    apply: view => {
                        getEditorConfig(view.state).onSubmit()
                    },
                    render: queryRenderer,
                },
            ],
        })
    }

    // Completions
    const completions = [...filterSuggestions(tokens, token, pos)]

    if (completions.length > 0) {
        result.push({ title: 'Narrow your search', options: completions.slice(0, 5) })
    }

    // Cached repos
    if (showRepoSuggestions) {
        const repos = cachedRepoSuggestions(token)
        if (repos.length > 0) {
            result.push({ title: 'Repositories', options: repos.slice(0, 5) })
        }
    }

    return {
        valid,
        result,
        next: showRepoSuggestions
            ? () =>
                  dynamicRepoSuggestions(token).then(suggestions => {
                      if (suggestions.length > 0) {
                          return {
                              result: [
                                  ...result.filter(group => group.title !== 'Repositories'),
                                  { title: 'Repositories', options: suggestions.slice(0, 5) },
                              ],
                              valid,
                          }
                      }
                      return { result, valid }
                  })
            : undefined,
    }
}
