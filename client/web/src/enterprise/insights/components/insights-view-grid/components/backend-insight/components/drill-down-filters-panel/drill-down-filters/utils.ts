import {
    Maybe,
    SeriesDisplayOptionsInput,
    SeriesSortDirection,
    SeriesSortMode,
} from '../../../../../../../../../graphql-operations'
import { MAX_NUMBER_OF_SERIES } from '../../../../../../../constants'
import { SeriesDisplayOptions, SeriesDisplayOptionsInputRequired } from '../../../../../../../core/types/insight/common'

import { DrillDownFiltersFormValues } from './DrillDownInsightFilters'

const DEFAULT_SERIES_DISPLAY_OPTIONS: SeriesDisplayOptionsInputRequired = {
    limit: 20,
    sortOptions: {
        direction: SeriesSortDirection.DESC,
        mode: SeriesSortMode.RESULT_COUNT,
    },
}

interface InsightRepositoriesFilter {
    include: string
    exclude: string
}

export function getSerializedRepositoriesFilter(filter: InsightRepositoriesFilter): string {
    const { include, exclude } = filter
    const includeString = include ? `repo:${include}` : ''
    const excludeString = exclude ? `-repo:${exclude}` : ''

    return `${includeString} ${excludeString}`.trim()
}

export const getSortPreview = (seriesDisplayOptions: {
    limit: string
    sortOptions: { direction: string; mode: string }
}): string => {
    const { sortOptions, limit } = seriesDisplayOptions

    const ascending = sortOptions.direction === SeriesSortDirection.ASC
    const mode = sortOptions.mode
    let sortBy

    switch (mode) {
        case undefined:
        case SeriesSortMode.RESULT_COUNT:
            sortBy = `by result count ${ascending ? 'low to high' : 'high to low'}`
            break
        case SeriesSortMode.LEXICOGRAPHICAL:
            sortBy = ascending ? 'A-Z' : 'Z-A'
            break
        case SeriesSortMode.DATE_ADDED:
            sortBy = `by date ${ascending ? 'newest to oldest' : 'oldest to newest'}`
            break
        default:
            sortBy = 'ERROR: Unknown sort type.'
            break
    }

    return `Sorted ${sortBy}, limit ${limit} series`
}

type InsightContextsFilter = string

export function getSerializedSearchContextFilter(
    filter: InsightContextsFilter,
    withContextPrefix: boolean = true
): string {
    const filterValue = filter !== '' ? filter : 'global (default)'

    return withContextPrefix ? `context:${filterValue}` : filterValue
}

/**
 * Returns a SeriesDisplayOptionsInput object with default values
 *
 * @param options series display options
 */
export const parseSeriesDisplayOptions = (
    options?: SeriesDisplayOptions | SeriesDisplayOptionsInput | DrillDownFiltersFormValues['seriesDisplayOptions']
): SeriesDisplayOptionsInputRequired => {
    if (!options) {
        return DEFAULT_SERIES_DISPLAY_OPTIONS
    }

    const limit = parseSeriesLimit(options?.limit) || MAX_NUMBER_OF_SERIES
    const sortOptions = options.sortOptions || DEFAULT_SERIES_DISPLAY_OPTIONS.sortOptions

    return {
        limit,
        sortOptions: {
            mode: sortOptions.mode || DEFAULT_SERIES_DISPLAY_OPTIONS.sortOptions.mode,
            direction: sortOptions.direction || DEFAULT_SERIES_DISPLAY_OPTIONS.sortOptions.direction,
        },
    }
}

export const parseSeriesLimit = (limit: string | Maybe<number> | undefined): number | undefined => {
    if (typeof limit === 'number') {
        return Math.min(limit, MAX_NUMBER_OF_SERIES)
    }

    if (!limit || limit.length === 0) {
        return
    }

    return Math.min(parseInt(limit, 10), MAX_NUMBER_OF_SERIES)
}
