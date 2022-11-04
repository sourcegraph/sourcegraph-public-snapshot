import { InsightDataNode } from '../../../../../../../graphql-operations'
import { BackendInsight, isComputeInsight } from '../../../../types'
import { InsightContentType } from '../../../../types/insight/common'
import { BackendInsightData } from '../../../code-insights-backend-types'
import { createComputeCategoricalChart } from '../../../utils/create-categorical-content'
import { createLineChartContent } from '../../../utils/create-line-chart-content'

export const createBackendInsightData = (insight: BackendInsight, response: InsightDataNode): BackendInsightData => {
    const seriesData = response.dataSeries
    const isFetchingHistoricalData = seriesData.some(({ status: { isLoadingData } }) => isLoadingData)
    // const percentComplete = seriesData.map(({ status: { percentComplete } }) => percentComplete)
    let pct = 0
    seriesData.forEach(({ status: { percentComplete } }) => {
        pct = percentComplete?.valueOf() || 0
    })

    if (isComputeInsight(insight)) {
        return {
            // We have to tweak original logic around historical data since compute powered
            // insights have problem with generated data series status info
            // see https://github.com/sourcegraph/sourcegraph/issues/38893
            isFetchingHistoricalData,
            percentComplete: pct,
            data: {
                type: InsightContentType.Categorical,
                content: createComputeCategoricalChart(insight, seriesData),
            },
        }
    }

    return {
        isFetchingHistoricalData,
        percentComplete: pct,
        data: {
            type: InsightContentType.Series,
            content: createLineChartContent(insight, seriesData),
        },
    }
}
