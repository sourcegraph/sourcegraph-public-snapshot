import { FilterTypes } from '../interactive/util'

export type SuggestionTypes = FilterTypes | NonFilterSuggestionType

/**
 * NonFilterSuggestionType represents the types of suggestion results that do not match a filter.
 *
 * For example, there is no `symbol:` filter, but there are symbol suggestion results.
 */
export enum NonFilterSuggestionType {
    filters = 'filters',
    dir = 'dir',
    symbol = 'symbol',
}
