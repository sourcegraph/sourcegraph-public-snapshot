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

        negated?: boolean
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
    // '-repo' = '-repo',
    // '-file' = '-file',
    // '-lang' = '-lang',
    // '-repohasfile' = '-repohasfile',
}

export const filterTypeKeys: FilterTypes[] = Object.keys(FilterTypes) as FilterTypes[]

export enum NegatedFilters {
    repo = '-repo',
    file = '-file',
    lang = '-lang',
    repohasfile = '-repohasfile',
}

/** The list of filters that are able to be negated. */
export type NegatableFilters = FilterTypes.repo | FilterTypes.file | FilterTypes.repohasfile | FilterTypes.lang

export const isNegatableFilter = (filter: FilterTypes): filter is NegatableFilters =>
    Object.keys(NegatedFilters).includes(filter)

/** The list of all negated filters. i.e. all valid filters that have `-` as a suffix. */
export const negatedFilters = Object.values(NegatedFilters)

export const isNegatedFilter = (filter: string): filter is NegatedFilters =>
    negatedFilters.includes(filter as NegatedFilters)

const negatedFilterToNegatableFilter: { [key: string]: NegatableFilters } = {
    '-repo': FilterTypes.repo,
    '-file': FilterTypes.file,
    '-lang': FilterTypes.lang,
    '-repohasfile': FilterTypes.repohasfile,
}

export const resolveNegatedFilter = (filter: NegatedFilters): NegatableFilters => negatedFilterToNegatableFilter[filter]
