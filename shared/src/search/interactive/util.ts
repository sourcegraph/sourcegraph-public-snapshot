import { SuggestionTypes } from '../suggestions/util'

/**
 * The data structure that holds the filters in a query.
 *
 * The data structureis a map, where the key is a uniquely assigned string in the form `repoType-numberOfFilterAdded`.
 * The value is a data structure containing the fields {`type`, `value`, `editable`}.
 * `type` is the field type of the filter (repo, file, etc.) `value` is the current value for that particular filter,
 *  and `editable` is whether the corresponding filter input is currently editable in the UI.
 */
export interface FiltersToTypeAndValue {
    [key: string]: { type: SuggestionTypes; value: string; editable: boolean }
}
