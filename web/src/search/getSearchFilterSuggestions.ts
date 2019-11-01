import { fetchSearchFilterSuggestions } from './backend'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { Suggestion, SuggestionTypes, FiltersSuggestionTypes } from './input/Suggestion'
import { mapValues } from 'lodash'
import { assign } from 'lodash/fp'

export type SearchFilterSuggestions = Record<
    FiltersSuggestionTypes,
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
        ].map(
            assign({
                type: SuggestionTypes.filters,
                label: 'add to query',
            })
        ),
    },
    type: {
        default: 'code',
        values: [{ title: 'code' }, { title: 'diff' }, { title: 'commit' }, { title: 'symbol' }].map(
            assign({
                type: SuggestionTypes.type,
            })
        ),
    },
    case: {
        default: 'no',
        values: [{ title: 'yes' }, { title: 'no' }].map(
            assign({
                type: SuggestionTypes.case,
            })
        ),
    },
    fork: {
        default: 'yes',
        values: [{ title: 'no' }, { title: 'only' }, { title: 'yes' }].map(
            assign({
                type: SuggestionTypes.fork,
            })
        ),
    },
    archived: {
        default: 'yes',
        values: [{ title: 'no' }, { title: 'only' }, { title: 'yes' }].map(
            assign({
                type: SuggestionTypes.archived,
            })
        ),
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
        ].map(
            assign({
                type: SuggestionTypes.file,
            })
        ),
    },
    lang: {
        values: [
            { title: 'c' },
            { title: 'cpp' },
            { title: 'csharp' },
            { title: 'css' },
            { title: 'go' },
            { title: 'haskell' },
            { title: 'html' },
            { title: 'java' },
            { title: 'javascript' },
            { title: 'lua' },
            { title: 'markdown' },
            { title: 'php' },
            { title: 'python' },
            { title: 'r' },
            { title: 'ruby' },
            { title: 'swift' },
            { title: 'typescript' },
        ].map(
            assign({
                type: SuggestionTypes.lang,
            })
        ),
    },
    repogroup: {
        values: [],
    },
    repo: {
        values: [],
    },
    repohasfile: {
        values: [{ title: 'go.mod' }, { title: 'package.json' }, { title: 'Gemfile' }].map(
            assign({
                type: SuggestionTypes.repohasfile,
            })
        ),
    },
    repohascommitafter: {
        values: [{ title: '1 week ago' }, { title: '1 month ago' }].map(
            assign({
                type: SuggestionTypes.repohascommitafter,
            })
        ),
    },
    count: {
        values: [{ title: '100' }, { title: '1000' }].map(
            assign({
                type: SuggestionTypes.count,
            })
        ),
    },
    timeout: {
        values: [{ title: '10s' }, { title: '30s' }].map(
            assign({
                type: SuggestionTypes.timeout,
            })
        ),
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
