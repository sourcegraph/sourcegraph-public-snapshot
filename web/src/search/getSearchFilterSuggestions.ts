import { fetchSearchFilterSuggestions } from './backend'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { Suggestion, SuggestionTypes } from './input/Suggestion'
import { mapValues } from 'lodash'

export type SearchFilterSuggestions = Record<
    SuggestionTypes,
    {
        default?: string
        values: Suggestion[]
    }
>

type PartialSearchFilterSuggestions = Record<
    SuggestionTypes,
    {
        default?: string
        values: Omit<Suggestion, 'type'>[]
    }
>

export const addTypeToSuggestions = (suggestions: PartialSearchFilterSuggestions): SearchFilterSuggestions =>
    Object.keys(suggestions).reduce<SearchFilterSuggestions>(
        (suggestionsWithTypes, type) => {
            const typeSuggestions = suggestions[type as SuggestionTypes]
            return {
                ...suggestionsWithTypes,
                [type]: {
                    ...typeSuggestions,
                    values: typeSuggestions.values.map(suggestion => ({ ...suggestion, type })),
                },
            }
        },
        suggestions as SearchFilterSuggestions
    )

export const filterAliases = {
    r: SuggestionTypes.repo,
    g: SuggestionTypes.repogroup,
    f: SuggestionTypes.file,
    l: SuggestionTypes.lang,
    language: SuggestionTypes.lang,
}

export const baseSuggestions: PartialSearchFilterSuggestions = {
    filters: {
        values: [
            {
                title: 'repo',
                description: 'regex-pattern (include results whose repository path matches)',
            },
            {
                title: '-repo',
                description: 'regex-pattern (exclude results whose repository path matches)',
            },
            {
                title: 'repogroup',
                description: 'group-name (include results from the named group)',
            },
            {
                title: 'repohasfile',
                description: 'regex-pattern (include results from repos that contain a matching file)',
            },
            {
                title: 'repohascommitafter',
                description: '"string specifying time frame" (filter out stale repositories without recent commits)',
            },
            {
                title: 'file',
                description: 'regex-pattern (include results whose file path matches)',
            },
            {
                title: '-file',
                description: 'regex-pattern (exclude results whose file path matches)',
            },
            {
                title: 'type',
                description: 'code | diff | commit | symbol',
            },
            {
                title: 'case',
                description: 'yes | no (default)',
            },
            {
                title: 'lang',
                description: 'lang-name (include results from the named language)',
            },
            {
                title: '-lang',
                description: 'lang-name (exclude results from the named language)',
            },
            {
                title: 'fork',
                description: 'no | only | yes (default)',
            },
            {
                title: 'archived',
                description: 'no | only | yes (default)',
            },
            {
                title: 'count',
                description: 'integer (number of results to fetch)',
            },
            {
                title: 'timeout',
                description: '"string specifying time duration" (duration before timeout)',
            },
        ],
    },
    type: {
        default: 'code',
        values: [{ title: 'code' }, { title: 'diff' }, { title: 'commit' }, { title: 'symbol' }],
    },
    case: {
        default: 'no',
        values: [{ title: 'yes' }, { title: 'no' }],
    },
    fork: {
        default: 'yes',
        values: [{ title: 'no' }, { title: 'only' }, { title: 'yes' }],
    },
    archived: {
        default: 'yes',
        values: [{ title: 'no' }, { title: 'only' }, { title: 'yes' }],
    },
    file: {
        values: [
            {
                title: '(test|spec)',
                description: 'Test files',
            },
            {
                title: '\\.json$',
                description: 'JSON files',
            },
            {
                title: '(vendor|node_modules)/',
                description: 'Vendored code',
            },
            {
                title: '\\.md$',
                description: 'Markdown files',
            },
            {
                title: '\\.(txt|md)$',
                description: 'Text documents',
            },
        ],
    },
    lang: {
        values: [{ title: 'javascript' }, { title: 'go' }, { title: 'markdown' }],
    },
    repogroup: {
        values: [],
    },
    repo: {
        values: [],
    },
    repohasfile: {
        values: [{ title: 'go.mod' }, { title: 'package.json' }, { title: 'Gemfile' }],
    },
    repohascommitafter: {
        values: [{ title: '1 week ago' }, { title: '1 month ago' }],
    },
    count: {
        values: [{ title: '100' }, { title: '1000' }],
    },
    timeout: {
        values: [{ title: '10s' }, { title: '30s' }],
    },
}

export const getSearchFilterSuggestions = (): Observable<SearchFilterSuggestions> =>
    fetchSearchFilterSuggestions().pipe(
        map(loadedSuggestions => {
            const { repo, repogroup } = loadedSuggestions
            const formattedValues = mapValues({ repo, repogroup }, values => ({
                values: values.map(title => ({ title })),
            }))
            return addTypeToSuggestions({
                ...baseSuggestions,
                ...formattedValues,
            })
        })
    )
