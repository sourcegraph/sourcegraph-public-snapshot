import { BackendInsightFilters } from '../../../../../../core/backend/types'
import {
    SearchBackendBasedInsightFiltersType,
    SearchBasedBackendFilters,
} from '../../../../../../core/types/insight/search-insight'

import { DrillDownFiltersFormValues } from './components/drill-down-filters-form/DrillDownFiltersForm'

export const EMPTY_DRILLDOWN_FILTERS: SearchBasedBackendFilters = {
    type: SearchBackendBasedInsightFiltersType.Regex,
    includeRepoRegexp: '',
    excludeRepoRegexp: '',
}

/**
 * Selector function from insight model filters to backend supported filters.
 */
export function getBackendFilters(filters: SearchBasedBackendFilters): BackendInsightFilters {
    // Currently we support only regexp filters so extract them in a separate object
    // to pass further in a gql api fetcher method
    if (filters.type === SearchBackendBasedInsightFiltersType.Regex) {
        return {
            excludeRepoRegexp: filters.excludeRepoRegexp.trim() ?? null,
            includeRepoRegexp: filters.includeRepoRegexp.trim() ?? null,
        }
    }

    // Fallback on empty backend filters
    return {
        excludeRepoRegexp: null,
        includeRepoRegexp: null,
    }
}

/**
 * Selector function from insight model filters to filter form values.
 */
export function getDrillDownFormValues(filters: SearchBasedBackendFilters): DrillDownFiltersFormValues {
    // Currently we support only regexp filters so extract them in a separate object
    // to pass further in a gql api fetcher method
    if (filters.type === SearchBackendBasedInsightFiltersType.Regex) {
        return {
            excludeRepoRegexp: filters.excludeRepoRegexp,
            includeRepoRegexp: filters.includeRepoRegexp,
        }
    }

    // Fallback on empty backend filters
    return {
        excludeRepoRegexp: '',
        includeRepoRegexp: '',
    }
}
