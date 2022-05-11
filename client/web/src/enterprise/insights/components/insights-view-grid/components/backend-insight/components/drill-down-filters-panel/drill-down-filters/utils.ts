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

type InsightContextsFilter = string

export function getSerializedSearchContextFilter(filter: InsightContextsFilter): string {
    return filter !== '' ? filter : 'global (default)'
}

// To simplify logic on the front end we ensure that a value is always proved
export const parseSeriesDisplayOptions = (options?: SeriesDisplayOptions): SeriesDisplayOptionsInputRequired => {
    if (!options) {
        return DEFAULT_SERIES_DISPLAY_OPTIONS
    }

    const limit = options.limit || DEFAULT_SERIES_DISPLAY_OPTIONS.limit
    const sortOptions = options.sortOptions || DEFAULT_SERIES_DISPLAY_OPTIONS.sortOptions

    return {
        limit,
        sortOptions: {
            mode: sortOptions.mode || DEFAULT_SERIES_DISPLAY_OPTIONS.sortOptions.mode,
            direction: sortOptions.direction || DEFAULT_SERIES_DISPLAY_OPTIONS.sortOptions.direction,
        },
    }
}
