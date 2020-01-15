/**
 * The data structure that holds the filters in a query.
 *
 */
export interface FiltersToTypeAndValue {
    /**
     * Key is a unique string of the form `filterType-numberOfFilterAdded`.
     * */
    [key: string]: {
        // `type` is the field type of the filter (repo, file, etc.)
        type: FilterTypes
        // `value` is the current value for that particular filter,
        value: string
        // `editable` is whether the corresponding filter input is currently editable in the UI.
        editable: boolean
    }
}

export enum FilterTypes {
    repo = 'repo',
    repogroup = 'repogroup',
    repohasfile = 'repohasfile',
    repohascommitafter = 'repohascommitafter',
    file = 'file',
    lang = 'lang',
    count = 'count',
    timeout = 'timeout',
    fork = 'fork',
    archived = 'archived',
    case = 'case',
}

export const filterTypeKeys: FilterTypes[] = Object.keys(FilterTypes) as FilterTypes[]
