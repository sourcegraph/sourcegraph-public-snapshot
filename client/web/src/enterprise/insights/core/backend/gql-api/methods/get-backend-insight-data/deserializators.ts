import { InsightDataNode, InsightDataSeries } from '../../../../../../../graphql-operations'
import { BackendInsight, InsightType } from '../../../../types'
import { SearchBasedInsightSeries } from '../../../../types/insight/search-insight'
import { BackendInsightData } from '../../../code-insights-backend-types'
import { createLineChartContentFromIndexedSeries } from '../../../utils/create-line-chart-content'

export const MAX_NUMBER_OF_SERIES = 20

export const createBackendInsightData = (insight: BackendInsight, response: InsightDataNode): BackendInsightData => {
    const rawSeries = response.dataSeries.slice(0, MAX_NUMBER_OF_SERIES)
    const series = getParsedSeries(insight, rawSeries)

    return {
        id: insight.id,
        view: {
            title: insight.title,
            content: [createLineChartContentFromIndexedSeries(rawSeries, series)],
            isFetchingHistoricalData: response.dataSeries.some(
                ({ status: { pendingJobs, backfillQueuedAt } }) => pendingJobs > 0 || backfillQueuedAt === null
            ),
        },
    }
}

const COLORS = ['grape', 'indigo', 'green', 'red', 'violet', 'lime', 'pink', 'blue', 'yellow', 'orange', 'cyan', 'teal']

const SERIES_COLORS = COLORS.map(name => `var(--oc-${name}-7)`)

function getParsedSeries(insight: BackendInsight, series: InsightDataSeries[]): SearchBasedInsightSeries[] {
    switch (insight.viewType) {
        case InsightType.SearchBased:
            return insight.series

        case InsightType.CaptureGroup: {
            const { query } = insight

            return series.map((generatedSeries, index) => ({
                id: generatedSeries.seriesId,
                name: generatedSeries.label,
                query,
                stroke: SERIES_COLORS[index % SERIES_COLORS.length],
            }))
        }
    }
}
