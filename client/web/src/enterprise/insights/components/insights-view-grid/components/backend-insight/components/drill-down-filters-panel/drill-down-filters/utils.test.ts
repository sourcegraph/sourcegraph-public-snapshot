import { SeriesSortDirection, SeriesSortMode } from '@sourcegraph/shared/src/graphql-operations'

import { SeriesDisplayOptionsInputRequired } from '../../../../../../../core/types/insight/common'

import { DrillDownFiltersFormValues } from './DrillDownInsightFilters'
import { parseSeriesDisplayOptions } from './utils'

const TEST_SERIES_DISPLAY_OPTIONS: DrillDownFiltersFormValues['seriesDisplayOptions'] = {
    limit: '10',
    sortOptions: {
        direction: SeriesSortDirection.ASC,
        mode: SeriesSortMode.DATE_ADDED,
    },
}

const PARSED_TEST_SERIES_DISPLAY_OPTIONS: SeriesDisplayOptionsInputRequired = {
    limit: 10,
    sortOptions: {
        direction: SeriesSortDirection.ASC,
        mode: SeriesSortMode.DATE_ADDED,
    },
}

describe('BackendInsight', () => {
    describe('parseSeriesDisplayOptions', () => {
        it('returns given object when provided complete values', () => {
            const parsed = parseSeriesDisplayOptions(TEST_SERIES_DISPLAY_OPTIONS)
            expect(parsed).toEqual(PARSED_TEST_SERIES_DISPLAY_OPTIONS)
        })
    })
})
