import { InsightDataNode, InsightDataSeries } from '../../../../../../../graphql-operations'
import { DATA_SERIES_COLORS } from '../../../../../pages/insights/creation/search-insight'
import { BackendInsight, InsightType, SearchBasedInsightSeries } from '../../../../types'
import { BackendInsightData } from '../../../code-insights-backend-types'
import { createLineChartContent } from '../../../utils/create-line-chart-content'

export const MAX_NUMBER_OF_SERIES = 20
export const DATA_SERIES_COLORS_LIST = Object.values(DATA_SERIES_COLORS)

export const createBackendInsightData = (insight: BackendInsight, response: InsightDataNode): BackendInsightData => {
    const seriesData = response.dataSeries.slice(0, MAX_NUMBER_OF_SERIES)
    const seriesMetadata = getParsedDataSeriesMetadata(insight, seriesData)

    return {
        content: createLineChartContent(seriesData, seriesMetadata, insight.filters),
        isFetchingHistoricalData: seriesData.some(
            ({ status: { pendingJobs, backfillQueuedAt } }) => pendingJobs > 0 || backfillQueuedAt === null
        ),
    }
}

function getParsedDataSeriesMetadata(
    insight: BackendInsight,
    seriesData: InsightDataSeries[]
): SearchBasedInsightSeries[] {
    switch (insight.type) {
        case InsightType.SearchBased:
            return insight.series

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
