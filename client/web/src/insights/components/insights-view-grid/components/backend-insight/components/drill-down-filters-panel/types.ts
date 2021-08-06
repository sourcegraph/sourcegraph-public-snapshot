export enum DrillDownFiltersMode {
    Regex = 'regex',
    Repolist = 'repolist',
}

export interface DrillDownFilters {
    mode: DrillDownFiltersMode
    includeRepoRegex: string
    excludeRepoRegex: string
}

export const EMPTY_DRILLDOWN_FILTERS = {
    mode: DrillDownFiltersMode.Regex,
    includeRepoRegex: '',
    excludeRepoRegex: '',
}
