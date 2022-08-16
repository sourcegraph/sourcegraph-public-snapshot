import { InsightDataNode } from '../../../../../../../graphql-operations'
import { BackendInsight, isComputeInsight } from '../../../../types'
import { InsightContentType } from '../../../../types/insight/common'
import { BackendInsightData } from '../../../code-insights-backend-types'
import { createComputeCategoricalChart } from '../../../utils/create-categorical-content'
import { createLineChartContent } from '../../../utils/create-line-chart-content'

export const MAX_NUMBER_OF_SERIES = 20

export const createBackendInsightData = (insight: BackendInsight, response: InsightDataNode): BackendInsightData => {
    const seriesData = response.dataSeries.slice(0, MAX_NUMBER_OF_SERIES)
    const isFetchingHistoricalData = seriesData.some(
        ({ status: { pendingJobs, backfillQueuedAt } }) => pendingJobs > 0 || backfillQueuedAt === null
    )

    if (isComputeInsight(insight)) {
        return {
            // We have to tweak original logic around historical data since compute powered
            // insights have problem with generated data series status info
            // see https://github.com/sourcegraph/sourcegraph/issues/38893
            isFetchingHistoricalData: isFetchingHistoricalData || seriesData.some(series => !series.label),
            data: {
                type: InsightContentType.Categorical,
                content: createComputeCategoricalChart(insight, seriesData),
            },
        }
    }

    return {
        isFetchingHistoricalData,
        data: {
            type: InsightContentType.Series,
            content: createLineChartContent(insight, seriesData),
        },
    }
}
