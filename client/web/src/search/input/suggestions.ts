import { extendedMatch, Fzf, FzfOptions } from 'fzf'
import { FilterDefinition, FILTERS, FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import {
    Group,
    Option,
    EntryOf,
    Source,
    FilterSuggestion,
} from '@sourcegraph/search-ui/src/input/codemirror/suggestions'
import { getDocumentNode, gql } from '@sourcegraph/http-client'
import { getWebGraphQLClient } from '../../backend/graphql'
import { SuggestionsRepoResult, SuggestionsRepoVariables } from 'src/graphql-operations'
import { tokenAt } from '@sourcegraph/search-ui/src/input/codemirror'
import { Token } from '@sourcegraph/shared/src/search/query/token'
import { mdiFilterOutline, mdiSourceRepository } from '@mdi/js'
import { EditorState } from '@codemirror/state'

const FILTER_SUGGESTIONS = new Fzf(
    Object.entries(FILTERS).map(
        ([name, definition]) => ({ name, definition } as { name: string; definition: FilterDefinition })
    ),
    { selector: item => item.name, match: extendedMatch }
)
const DEFAULT_FILTERS = [FilterType.repo, FilterType.type].map(name => ({ name, definition: FILTERS[name] }))

export const filterSuggestions = (token: Token | undefined, position: number): Option[] => {
    if (!token || token.type === 'whitespace') {
        return DEFAULT_FILTERS.map(entry => {
            const description =
                typeof entry.definition.description === 'function'
                    ? entry.definition.description(false)
                    : entry.definition.description

            return {
                type: 'completion',
                icon: mdiFilterOutline,
                render: FilterSuggestion,
                value: entry.name,
                insertValue: entry.name + ':',
                description,
                from: position,
            }
        })
    }

    if (token?.type === 'pattern') {
        return FILTER_SUGGESTIONS.find('^' + token.value).map(entry => {
            const description =
                typeof entry.item.definition.description === 'function'
                    ? entry.item.definition.description(false)
                    : entry.item.definition.description

            return {
                type: 'completion',
                icon: mdiFilterOutline,
                render: FilterSuggestion,
                value: entry.item.name,
                insertValue: entry.item.name + ':',
                description,
                from: token.range.start,
                to: token.range.end,
                matches: entry.positions,
            }
        })
    }
    return []
}

export const staticFilterValueSuggestions = (token?: Token): Option[] => {
    if (token?.type !== 'filter') {
        return []
    }

    const resolvedFilter = resolveFilter(token.field.value)
    if (!resolvedFilter?.definition.discreteValues) {
        return []
    }

    const value = token.value
    const options: EntryOf<'completion'>[] = resolvedFilter.definition
        .discreteValues(token.value, false)
        .map(value => ({
            type: 'completion',
            from: token.value?.range.start ?? token.range.end,
            to: token.value?.range.end,
            value: value.label,
            insertValue: (value.insertText ?? value.label) + ' ',
        }))
    if (!value || value.value === '') {
        return options
    }

    const fzf = new Fzf(options, { selector: option => option.value })
    return fzf.find(value.value).map(match => ({ ...match.item, matches: match.positions }))
}

function starTiebraker(a: { item: { stars: number } }, b: { item: { stars: number } }): number {
    return b.item.stars - a.item.stars
}

const repoFzfOptions: FzfOptions<EntryOf<'target'> & { stars: number }> = {
    selector: item => item.value,
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

function repoMatches(query: string, repos: Repo[]): (EntryOf<'target'> & { stars: number })[] {
    const queryResults: (EntryOf<'target'> & { stars: number })[] = repos.map(({ name, stars }) => ({
        type: 'target',
        icon: mdiSourceRepository,
        value: name,
        url: `/${name}`,
        stars,
    }))
    const fzf = new Fzf(queryResults, repoFzfOptions)
    return fzf.find(query).map(match => ({ ...match.item, matches: match.positions }))
}

export const dynamicRepoSuggestions = async (token?: Token) => {
    if (token?.type !== 'pattern') {
        return []
    }

    const query = `type:repo count:50 ${token.value}`
    const repositories =
        queryCache.get(query) ??
        getWebGraphQLClient()
            .then(client =>
                client.query<SuggestionsRepoResult, SuggestionsRepoVariables>({
                    query: getDocumentNode(REPOS_QUERY),
                    variables: { query: `type:repo count:50 ${token.value}` },
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
    return repoMatches(token.value, [...repoCache.values()])
}

const cachedRepoSuggestions = (token?: Token) => {
    if (token?.type !== 'pattern') {
        return []
    }

    return repoMatches(token.value, [...repoCache.values()])
}

export const suggestions: Source = (state, pos) => {
    const token = tokenAt(state, pos)
    const result: Group[] = []

    // Completions
    const completions = [...filterSuggestions(token, pos), ...staticFilterValueSuggestions(token)]

    if (completions.length > 0) {
        result.push({ title: 'Narrow your search', entries: completions.slice(0, 5) })
    }

    // Cached repos
    const repos = cachedRepoSuggestions(token)
    if (repos.length > 0) {
        result.push({ title: 'Repositories', entries: repos.slice(0, 5) })
    }

    function valid(state: EditorState, position: number) {
        return token === tokenAt(state, position)
    }

    return {
        result: result,
        asyncResult: dynamicRepoSuggestions(token).then(suggestions => {
            if (suggestions.length > 0) {
                return {
                    result: [
                        ...result.filter(group => group.title !== 'Repositories'),
                        { title: 'Repositories', entries: suggestions.slice(0, 5) },
                    ],
                    valid,
                }
            }
            return { result }
        }),
        valid,
    }
}
