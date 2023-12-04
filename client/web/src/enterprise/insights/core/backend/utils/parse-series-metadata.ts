import type { InsightDataSeries } from '../../../../../graphql-operations'
import { DATA_SERIES_COLORS_LIST } from '../../../constants'
import { type BackendInsight, InsightType, type SearchBasedInsightSeries } from '../../types'

export function getParsedSeriesMetadata(
    insight: BackendInsight,
    seriesData: InsightDataSeries[]
): SearchBasedInsightSeries[] {
    switch (insight.type) {
        case InsightType.SearchBased: {
            return insight.series
        }

        case InsightType.Compute: {
            return seriesData.map((generatedSeries, index) => ({
                id: generatedSeries.seriesId,
                name: generatedSeries.label,
                // TODO we don't know compute series contributions to each data items in dataset
                // see https://github.com/sourcegraph/sourcegraph/issues/38832
                query: '',
                stroke: DATA_SERIES_COLORS_LIST[index % DATA_SERIES_COLORS_LIST.length],
            }))
        }

        case InsightType.CaptureGroup: {
            const { query } = insight

            return seriesData.map((generatedSeries, index) => ({
                id: generatedSeries.seriesId,
                name: generatedSeries.label,
                query,
                stroke: DATA_SERIES_COLORS_LIST[index % DATA_SERIES_COLORS_LIST.length],
            }))
        }
    }
}
