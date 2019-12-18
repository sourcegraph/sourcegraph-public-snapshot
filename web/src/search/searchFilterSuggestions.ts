import { Suggestion, FiltersSuggestionTypes } from './input/Suggestion'
import { assign } from 'lodash/fp'
import { languageIcons } from '../../../shared/src/components/languageIcons'
import { SuggestionTypes } from '../../../shared/src/search/suggestions/util'

export type SearchFilterSuggestions = Record<
    FiltersSuggestionTypes,
    {
        default?: string
        values: Suggestion[]
    }
>

export const searchFilterSuggestions: SearchFilterSuggestions = {
    filters: {
        values: [
            {
                value: 'repo:',
                description: 'regex-pattern (include results whose repository path matches)',
            },
            {
                value: '-repo:',
                description: 'regex-pattern (exclude results whose repository path matches)',
            },
            {
                value: 'repogroup:',
                description: 'group-name (include results from the named group)',
            },
            {
                value: 'repohasfile:',
                description: 'regex-pattern (include results from repos that contain a matching file)',
            },
            {
                value: '-repohasfile:',
                description: 'regex-pattern (exclude results from repositories that contain a matching file)',
            },
            {
                value: 'repohascommitafter:',
                description: '"string specifying time frame" (filter out stale repositories without recent commits)',
            },
            {
                value: 'file:',
                description: 'regex-pattern (include results whose file path matches)',
            },
            {
                value: '-file:',
                description: 'regex-pattern (exclude results whose file path matches)',
            },
            {
                value: 'type:',
                description: 'code | diff | commit | symbol',
            },
            {
                value: 'case:',
                description: 'yes | no (default)',
            },
            {
                value: 'lang:',
                description: 'lang-name (include results from the named language)',
            },
            {
                value: '-lang:',
                description: 'lang-name (exclude results from the named language)',
            },
            {
                value: 'fork:',
                description: 'no | only | yes (default)',
            },
            {
                value: 'archived:',
                description: 'no | only | yes (default)',
            },
            {
                value: 'count:',
                description: 'integer (number of results to fetch)',
            },
            {
                value: 'timeout:',
                description: '"string specifying time duration" (duration before timeout)',
            },
        ].map(
            assign({
                type: SuggestionTypes.filters,
            })
        ),
    },
    type: {
        default: 'code',
        values: [{ value: 'code' }, { value: 'diff' }, { value: 'commit' }, { value: 'symbol' }].map(
            assign({
                type: SuggestionTypes.type,
            })
        ),
    },
    case: {
        default: 'no',
        values: [{ value: 'yes' }, { value: 'no' }].map(
            assign({
                type: SuggestionTypes.case,
            })
        ),
    },
    fork: {
        default: 'yes',
        values: [{ value: 'no' }, { value: 'only' }, { value: 'yes' }].map(
            assign({
                type: SuggestionTypes.fork,
            })
        ),
    },
    archived: {
        default: 'yes',
        values: [{ value: 'no' }, { value: 'only' }, { value: 'yes' }].map(
            assign({
                type: SuggestionTypes.archived,
            })
        ),
    },
    file: {
        values: [
            { value: '(test|spec)', displayValue: 'Test files' },
            { value: '.(txt|md)', displayValue: 'Text files' },
        ].map(suggestion => ({
            ...suggestion,
            description: suggestion.value,
            type: SuggestionTypes.file,
        })),
    },
    lang: {
        values: Object.keys(languageIcons).map(value => ({ type: SuggestionTypes.lang, value })),
    },
    repogroup: {
        values: [],
    },
    repo: {
        values: [],
    },
    repohasfile: {
        values: [],
    },
    repohascommitafter: {
        values: [{ value: "'1 week ago'" }, { value: "'1 month ago'" }].map(
            assign({
                type: SuggestionTypes.repohascommitafter,
            })
        ),
    },
    count: {
        values: [{ value: '100' }, { value: '1000' }].map(
            assign({
                type: SuggestionTypes.count,
            })
        ),
    },
    timeout: {
        values: [{ value: '10s' }, { value: '30s' }].map(
            assign({
                type: SuggestionTypes.timeout,
            })
        ),
    },
}
