import { InsightDataNode } from '../../../../../../../graphql-operations'
import { BackendInsight, isComputeInsight } from '../../../../types'
import { InsightContentType } from '../../../../types/insight/common'
import { BackendInsightData } from '../../../code-insights-backend-types'
import { createComputeCategoricalChart } from '../../../utils/create-categorical-content'
import { createLineChartContent } from '../../../utils/create-line-chart-content'

export const createBackendInsightData = (insight: BackendInsight, response: InsightDataNode): BackendInsightData => {
    const seriesData = response.dataSeries
    const isFetchingHistoricalData = seriesData.some(({ status: { isLoadingData } }) => isLoadingData)
    const isAllSeriesErrored = seriesData.every(series => series.status.incompleteDatapoints.length > 0)

    if (isComputeInsight(insight)) {
        return {
            isFetchingHistoricalData,
            isAllSeriesErrored,
            data: {
                type: InsightContentType.Categorical,
                content: createComputeCategoricalChart(insight, seriesData),
            },
        }
    }

    return {
        isFetchingHistoricalData,
        isAllSeriesErrored,
        data: {
            type: InsightContentType.Series,
            content: createLineChartContent(insight, seriesData, !isAllSeriesErrored),
        },
    }
}
