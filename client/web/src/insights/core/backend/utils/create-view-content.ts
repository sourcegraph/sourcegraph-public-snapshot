import { LineChartContent } from 'sourcegraph'

import { InsightFields } from '../../../../graphql-operations'

export function createViewContent(
    insight: InsightFields
): LineChartContent<{ dateTime: number; [seriesKey: string]: number }, 'dateTime'> {
    const dataByXValue = new Map<string, { dateTime: number; [seriesKey: string]: number }>()
    for (const [seriesIndex, series] of insight.series.entries()) {
        for (const point of series.points) {
            let dataObject = dataByXValue.get(point.dateTime)
            if (!dataObject) {
                dataObject = {
                    dateTime: Date.parse(point.dateTime),
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
            name: series.label,
            dataKey: `series${index}`,
        })),
        xAxis: {
            dataKey: 'dateTime',
            scale: 'time',
            type: 'number',
        },
    }
}
