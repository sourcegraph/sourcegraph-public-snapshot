import { SeriesSortDirection, SeriesSortMode } from '../../../../../../../../../graphql-operations'
import { DEFAULT_SERIES_DISPLAY_OPTIONS } from '../../../../../../../core'
import { SeriesDisplayOptionsInputRequired } from '../../../../../../../core/types/insight/common'

import { parseSeriesDisplayOptions } from './utils'

const TEST_SERIES_DISPLAY_OPTIONS: SeriesDisplayOptionsInputRequired = {
    limit: 10,
    sortOptions: {
        direction: SeriesSortDirection.ASC,
        mode: SeriesSortMode.DATE_ADDED,
    },
}

describe('BackendInsight', () => {
    describe('parseSeriesDisplayOptions', () => {
        it('returns given object when provided complete values', () => {
            const parsed = parseSeriesDisplayOptions(TEST_SERIES_DISPLAY_OPTIONS.limit, TEST_SERIES_DISPLAY_OPTIONS)
            expect(parsed).toEqual(TEST_SERIES_DISPLAY_OPTIONS)
        })

        it('provides default limit', () => {
            const parsed = parseSeriesDisplayOptions(TEST_SERIES_DISPLAY_OPTIONS.limit, {
                ...TEST_SERIES_DISPLAY_OPTIONS,
                limit: null,
            })
            expect(parsed.limit).toEqual(TEST_SERIES_DISPLAY_OPTIONS.limit)
        })

        it('provides default sortOptions', () => {
            const parsed = parseSeriesDisplayOptions(TEST_SERIES_DISPLAY_OPTIONS.limit, {
                ...TEST_SERIES_DISPLAY_OPTIONS,
                sortOptions: { mode: null, direction: null },
            })
            expect(parsed.sortOptions).toEqual(DEFAULT_SERIES_DISPLAY_OPTIONS.sortOptions)
        })
    })
})
