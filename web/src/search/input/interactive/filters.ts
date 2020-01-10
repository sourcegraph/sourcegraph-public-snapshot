import { SuggestionTypes } from '../../../../../shared/src/search/suggestions/util'
import { Suggestion } from '../Suggestion'
import { assign } from 'lodash/fp'

export enum FilterTypes {
    repo = 'repo',
    repogroup = 'repogroup',
    repohasfile = 'repohasfile',
    repohascommitafter = 'repohascommitafter',
    file = 'file',
    case = 'case',
    lang = 'lang',
    fork = 'fork',
    archived = 'archived',
    count = 'count',
    timeout = 'timeout',
    dir = 'dir',
    symbol = 'symbol',
}

export const filterTypeKeys = Object.keys(FilterTypes)

export type textFilters = Exclude<FilterTypes, 'archived' | 'fork' | 'case'>

export type finiteFilterTypes = SuggestionTypes.archived | SuggestionTypes.fork

// TODO: Remove SuggestionTypes
export function isTextFilter(filter: FilterTypes | SuggestionTypes): boolean {
    const validTextFilters = [
        'repo',
        'repogroup',
        'repohasfile',
        'repohascommitafter',
        'file',
        'lang',
        'count',
        'timeout',
        'dir',
        'symbol',
    ]

    return validTextFilters.includes(filter)
}

export const finiteFilters: Record<
    finiteFilterTypes,
    {
        default: string
        values: Suggestion[]
    }
> = {
    archived: {
        default: 'yes',
        values: [{ value: 'no' }, { value: 'only' }, { value: 'yes' }].map(
            assign({
                type: SuggestionTypes.fork,
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
}

export const isFiniteFilter = (filter: SuggestionTypes): filter is finiteFilterTypes =>
    ['archived', 'fork', 'case'].includes(filter)
