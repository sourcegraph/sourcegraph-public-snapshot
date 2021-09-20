import { LineChartContent } from 'sourcegraph'

import { InsightFields } from '../../../../../graphql-operations'
import { SearchBasedInsightSeries } from '../../types/insight/search-insight'

export function createViewContent(
    insight: InsightFields,
    seriesSettings: SearchBasedInsightSeries[] = []
): LineChartContent<{ dateTime: number; [seriesKey: string]: number }, 'dateTime'> {
    // Immutable sort is required to avoid breaking useCallback memoziation in BackendInsight component
    const sortedSeriesSettings = [...seriesSettings].sort((a, b) => a.query.localeCompare(b.query))
    const dataByXValue = new Map<string, { dateTime: number; [seriesKey: string]: number }>()

    for (const [seriesIndex, series] of insight.series.entries()) {
        for (const point of series.points) {
            let dataObject = dataByXValue.get(point.dateTime)
            if (!dataObject) {
                dataObject = {
                    dateTime: Date.parse(point.dateTime),
                    // Initialize all series to null (empty chart) value
                    ...Object.fromEntries(insight.series.map((series, index) => [`series${index}`, null])),
                }
                dataByXValue.set(point.dateTime, dataObject)
            }
            dataObject[`series${seriesIndex}`] = point.value
        }
    }

    return {
        chart: 'line',
        data: [...dataByXValue.values()],
        series: insight.series.map((series, index) => ({
            name: sortedSeriesSettings[index]?.name ?? series.label,
            dataKey: `series${index}`,
            stroke: sortedSeriesSettings[index]?.stroke,
        })),
        xAxis: {
            dataKey: 'dateTime',
            scale: 'time',
            type: 'number',
        },
    }
}
