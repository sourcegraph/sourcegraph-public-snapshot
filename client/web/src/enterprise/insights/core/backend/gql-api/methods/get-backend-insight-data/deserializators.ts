import { InsightDataNode } from '../../../../../../../graphql-operations'
import { BackendInsight, InsightType } from '../../../../types'
import { SearchBasedInsightSeries } from '../../../../types/insight/search-insight'
import { BackendInsightData } from '../../../code-insights-backend-types'
import { createLineChartContentFromIndexedSeries } from '../../../utils/create-line-chart-content'

export const createBackendInsightData = (insight: BackendInsight, response: InsightDataNode): BackendInsightData => {
    const series = getParsedSeries(insight, response)

    return {
        id: insight.id,
        view: {
            title: insight.title,
            content: [createLineChartContentFromIndexedSeries(response.dataSeries, series)],
            isFetchingHistoricalData: response.dataSeries.some(
                ({ status: { pendingJobs, backfillQueuedAt } }) => pendingJobs > 0 || backfillQueuedAt === null
            ),
        },
    }
}

const COLORS = ['grape', 'indigo', 'green', 'red', 'violet', 'lime', 'pink', 'blue', 'yellow', 'orange', 'cyan', 'teal']

const SERIES_COLORS = COLORS.map(name => `var(--oc-${name}-7)`)

function getParsedSeries(insight: BackendInsight, response: InsightDataNode): SearchBasedInsightSeries[] {
    switch (insight.viewType) {
        case InsightType.SearchBased:
            return insight.series

        case InsightType.CaptureGroup: {
            const { query } = insight

            return response.dataSeries.map((generatedSeries, index) => ({
                id: generatedSeries.seriesId,
                name: generatedSeries.label,
                query,
                stroke: SERIES_COLORS[index % SERIES_COLORS.length],
            }))
        }
    }
}
