import { SeriesSortDirection, SeriesSortMode } from '../../../../../../../../../graphql-operations'
import type { InsightSeriesDisplayOptions } from '../../../../../../../core/types/insight/common'

export const getSerializedSortAndLimitFilter = (seriesDisplayOptions: InsightSeriesDisplayOptions): string => {
    const { sortOptions, limit, numSamples } = seriesDisplayOptions

    const ascending = sortOptions.direction === SeriesSortDirection.ASC
    const mode = sortOptions.mode
    let sortBy

    switch (mode) {
        case undefined:
        case SeriesSortMode.RESULT_COUNT: {
            sortBy = `by result count ${ascending ? 'low to high' : 'high to low'}`
            break
        }
        case SeriesSortMode.LEXICOGRAPHICAL: {
            sortBy = ascending ? 'A-Z' : 'Z-A'
            break
        }
        case SeriesSortMode.DATE_ADDED: {
            sortBy = `by date ${ascending ? 'newest to oldest' : 'oldest to newest'}`
            break
        }
        default: {
            sortBy = 'ERROR: Unknown sort type.'
            break
        }
    }

    return `Sorted ${sortBy}, limit ${limit ?? 20} series, max point per series ${numSamples ?? 90}`
}

type InsightContextsFilter = string

export function getSerializedSearchContextFilter(
    filter: InsightContextsFilter,
    withContextPrefix: boolean = true
): string {
    const filterValue = filter !== '' ? filter : 'global (default)'

    return withContextPrefix ? `context:${filterValue}` : filterValue
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
