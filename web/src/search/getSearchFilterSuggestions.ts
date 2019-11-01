import { fetchSearchFilterSuggestions } from './backend'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { Suggestion, SuggestionTypes } from './input/Suggestion'
import { mapValues } from 'lodash'

export type SearchFilterSuggestions = Record<
    Exclude<SuggestionTypes, SuggestionTypes.dir | SuggestionTypes.symbol>,
    {
        default?: string
        values: Suggestion[]
    }
>

export const filterAliases = {
    r: SuggestionTypes.repo,
    g: SuggestionTypes.repogroup,
    f: SuggestionTypes.file,
    l: SuggestionTypes.lang,
    language: SuggestionTypes.lang,
}

export const filterSuggestions: SearchFilterSuggestions = {
    filters: {
        values: [
            {
                type: SuggestionTypes.filters,
                title: 'repo',
                description: 'regex-pattern (include results whose repository path matches)',
            },
            {
                type: SuggestionTypes.filters,
                title: '-repo',
                description: 'regex-pattern (exclude results whose repository path matches)',
            },
            {
                type: SuggestionTypes.filters,
                title: 'repogroup',
                description: 'group-name (include results from the named group)',
            },
            {
                type: SuggestionTypes.filters,
                title: 'repohasfile',
                description: 'regex-pattern (include results from repos that contain a matching file)',
            },
            {
                type: SuggestionTypes.filters,
                title: 'repohascommitafter',
                description: '"string specifying time frame" (filter out stale repositories without recent commits)',
            },
            {
                type: SuggestionTypes.filters,
                title: 'file',
                description: 'regex-pattern (include results whose file path matches)',
            },
            {
                type: SuggestionTypes.filters,
                title: '-file',
                description: 'regex-pattern (exclude results whose file path matches)',
            },
            {
                type: SuggestionTypes.filters,
                title: 'type',
                description: 'code | diff | commit | symbol',
            },
            {
                type: SuggestionTypes.filters,
                title: 'case',
                description: 'yes | no (default)',
            },
            {
                type: SuggestionTypes.filters,
                title: 'lang',
                description: 'lang-name (include results from the named language)',
            },
            {
                type: SuggestionTypes.filters,
                title: '-lang',
                description: 'lang-name (exclude results from the named language)',
            },
            {
                type: SuggestionTypes.filters,
                title: 'fork',
                description: 'no | only | yes (default)',
            },
            {
                type: SuggestionTypes.filters,
                title: 'archived',
                description: 'no | only | yes (default)',
            },
            {
                type: SuggestionTypes.filters,
                title: 'count',
                description: 'integer (number of results to fetch)',
            },
            {
                type: SuggestionTypes.filters,
                title: 'timeout',
                description: '"string specifying time duration" (duration before timeout)',
            },
        ],
    },
    type: {
        default: 'code',
        values: [
            { type: SuggestionTypes.type, title: 'code' },
            { type: SuggestionTypes.type, title: 'diff' },
            { type: SuggestionTypes.type, title: 'commit' },
            { type: SuggestionTypes.type, title: 'symbol' },
        ],
    },
    case: {
        default: 'no',
        values: [{ type: SuggestionTypes.case, title: 'yes' }, { type: SuggestionTypes.case, title: 'no' }],
    },
    fork: {
        default: 'yes',
        values: [
            { type: SuggestionTypes.fork, title: 'no' },
            { type: SuggestionTypes.fork, title: 'only' },
            { type: SuggestionTypes.fork, title: 'yes' },
        ],
    },
    archived: {
        default: 'yes',
        values: [
            { type: SuggestionTypes.archived, title: 'no' },
            { type: SuggestionTypes.archived, title: 'only' },
            { type: SuggestionTypes.archived, title: 'yes' },
        ],
    },
    file: {
        values: [
            {
                type: SuggestionTypes.file,
                title: '(test|spec)',
                description: 'Test files',
            },
            {
                type: SuggestionTypes.file,
                title: '\\.json$',
                description: 'JSON files',
            },
            {
                type: SuggestionTypes.file,
                title: '(vendor|node_modules)/',
                description: 'Vendored code',
            },
            {
                type: SuggestionTypes.file,
                title: '\\.md$',
                description: 'Markdown files',
            },
            {
                type: SuggestionTypes.file,
                title: '\\.(txt|md)$',
                description: 'Text documents',
            },
        ],
    },
    lang: {
        values: [
            { type: SuggestionTypes.lang, title: 'c' },
            { type: SuggestionTypes.lang, title: 'cpp' },
            { type: SuggestionTypes.lang, title: 'csharp' },
            { type: SuggestionTypes.lang, title: 'css' },
            { type: SuggestionTypes.lang, title: 'go' },
            { type: SuggestionTypes.lang, title: 'haskell' },
            { type: SuggestionTypes.lang, title: 'html' },
            { type: SuggestionTypes.lang, title: 'java' },
            { type: SuggestionTypes.lang, title: 'javascript' },
            { type: SuggestionTypes.lang, title: 'lua' },
            { type: SuggestionTypes.lang, title: 'markdown' },
            { type: SuggestionTypes.lang, title: 'php' },
            { type: SuggestionTypes.lang, title: 'python' },
            { type: SuggestionTypes.lang, title: 'r' },
            { type: SuggestionTypes.lang, title: 'ruby' },
            { type: SuggestionTypes.lang, title: 'swift' },
            { type: SuggestionTypes.lang, title: 'typescript' },
        ],
    },
    repogroup: {
        values: [],
    },
    repo: {
        values: [],
    },
    repohasfile: {
        values: [
            { type: SuggestionTypes.repohasfile, title: 'go.mod' },
            { type: SuggestionTypes.repohasfile, title: 'package.json' },
            { type: SuggestionTypes.repohasfile, title: 'Gemfile' },
        ],
    },
    repohascommitafter: {
        values: [
            { type: SuggestionTypes.repohascommitafter, title: '1 week ago' },
            { type: SuggestionTypes.repohascommitafter, title: '1 month ago' },
        ],
    },
    count: {
        values: [{ type: SuggestionTypes.count, title: '100' }, { type: SuggestionTypes.count, title: '1000' }],
    },
    timeout: {
        values: [{ type: SuggestionTypes.timeout, title: '10s' }, { type: SuggestionTypes.timeout, title: '30s' }],
    },
}

/**
 * Fetch filter suggestions that are dynamic, currently repo and repogroup.
 */
export const getSearchFilterSuggestions = (): Observable<SearchFilterSuggestions> =>
    fetchSearchFilterSuggestions().pipe(
        map(loadedSuggestions => {
            const { repo, repogroup } = loadedSuggestions

            const fetchedFilterSuggestions = mapValues({ repo, repogroup }, (values, type: SuggestionTypes) => ({
                values: values.map(title => ({ title, type })),
            }))

            return {
                ...filterSuggestions,
                ...fetchedFilterSuggestions,
            }
        })
    )
