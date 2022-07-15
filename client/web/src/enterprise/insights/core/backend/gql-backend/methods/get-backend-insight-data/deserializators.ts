import { InsightDataNode } from '../../../../../../../graphql-operations'
import { DATA_SERIES_COLORS } from '../../../../../constants'
import { BackendInsight } from '../../../../types'
import { BackendInsightData } from '../../../code-insights-backend-types'
import { createLineChartContent } from '../../../utils/create-line-chart-content'

export const MAX_NUMBER_OF_SERIES = 20
export const DATA_SERIES_COLORS_LIST = Object.values(DATA_SERIES_COLORS)

export const createBackendInsightData = (insight: BackendInsight, response: InsightDataNode): BackendInsightData => {
    const seriesData = response.dataSeries.slice(0, MAX_NUMBER_OF_SERIES)

    return {
        content: createLineChartContent(insight, seriesData),
        isFetchingHistoricalData: seriesData.some(
            ({ status: { pendingJobs, backfillQueuedAt } }) => pendingJobs > 0 || backfillQueuedAt === null
        ),
    }
}
