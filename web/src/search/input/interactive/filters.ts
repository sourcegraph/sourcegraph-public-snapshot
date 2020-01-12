import { SuggestionTypes } from '../../../../../shared/src/search/suggestions/util'
import { Suggestion } from '../Suggestion'
import { assign } from 'lodash/fp'
import { FilterTypes } from '../../../../../shared/src/search/interactive/util'

export type textFilters = Exclude<FilterTypes, 'archived' | 'fork' | 'case'>

export type finiteFilterTypes = FilterTypes.archived | FilterTypes.fork

export function isTextFilter(filter: FilterTypes): boolean {
    const validTextFilters = [
        'repo',
        'repogroup',
        'repohasfile',
        'repohascommitafter',
        'file',
        'lang',
        'count',
        'timeout',
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

export const isFiniteFilter = (filter: FilterTypes): filter is finiteFilterTypes =>
    ['archived', 'fork'].includes(filter)

export const isolatedFuzzySearchFiltersFilterType = [FilterTypes.repo, FilterTypes.repogroup]
