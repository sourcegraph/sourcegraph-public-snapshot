import {
    SeriesDisplayOptionsInput,
    SeriesSortDirection,
    SeriesSortMode,
} from '../../../../../../../../../graphql-operations'
import { DEFAULT_SERIES_DISPLAY_OPTIONS } from '../../../../../../../core'
import { SeriesDisplayOptions, SeriesDisplayOptionsInputRequired } from '../../../../../../../core/types/insight/common'
import { Validator } from '../../../../../../form/hooks/useField'

export const validRegexp: Validator<string> = (value = '') => {
    if (value.trim() === '') {
        return
    }

    try {
        new RegExp(value)

        return
    } catch {
        return 'Must be a valid regular expression string'
    }
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

export const getSortPreview = (seriesDisplayOptions: SeriesDisplayOptionsInput): string => {
    const { sortOptions, limit } = seriesDisplayOptions

    if (!limit) {
        throw new Error('Limit is required to parse series display options.')
    }

    const ascending = sortOptions?.direction === SeriesSortDirection.ASC
    const mode = sortOptions?.mode
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
 * @param seriesCount The total series available for this insight. Used to set the max limit value
 * @param options series display options
 */
export const parseSeriesDisplayOptions = (
    seriesCount: number,
    options?: SeriesDisplayOptions | SeriesDisplayOptionsInput
): SeriesDisplayOptionsInputRequired => {
    if (!options) {
        return { ...DEFAULT_SERIES_DISPLAY_OPTIONS, limit: seriesCount }
    }

    const limit = Math.min(options.limit || seriesCount, seriesCount)
    const sortOptions = options.sortOptions || DEFAULT_SERIES_DISPLAY_OPTIONS.sortOptions

    return {
        limit,
        sortOptions: {
            mode: sortOptions.mode || DEFAULT_SERIES_DISPLAY_OPTIONS.sortOptions.mode,
            direction: sortOptions.direction || DEFAULT_SERIES_DISPLAY_OPTIONS.sortOptions.direction,
        },
    }
}
