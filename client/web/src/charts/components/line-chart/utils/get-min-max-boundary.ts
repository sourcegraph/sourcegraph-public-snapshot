import { LineChartSeriesWithData } from './data-series-processing'

interface MinMaxBoundariesInput<D> {
    dataSeries: LineChartSeriesWithData<D>[]
    xAxisKey: keyof D
}

interface Boundaries {
    minX: number
    minY: number
    maxX: number
    maxY: number
}

/**
 * Calculates min/max ranges for Y (across all available data series) and X axis
 * (time interval) global for all lines on the chart.
 */
export function getMinMaxBoundaries<D>(props: MinMaxBoundariesInput<D>): Boundaries {
    const { dataSeries, xAxisKey } = props

    let minX
    let maxX
    let minY
    let maxY

    for (const line of dataSeries) {
        for (const datum of line.data) {
            minX = Math.min(+datum[xAxisKey], minX ?? +datum[xAxisKey])
            maxX = Math.max(+datum[xAxisKey], maxX ?? +datum[xAxisKey])

            minY = Math.min(+datum[line.dataKey], minY ?? +datum[line.dataKey])
            maxY = Math.max(+datum[line.dataKey], maxY ?? +datum[line.dataKey])
        }
    }

    ;[minY, maxY, minX, maxX] = [minY ?? 0, maxY ?? 0, minX ?? 0, maxX ?? 0]

    // Expand range for better ticks looking in case if we got a flat data series dataset
    ;[minY, maxY] = minY === maxY ? [maxY - maxY / 2, maxY + maxY / 2] : [minY, maxY]

    return { minX, minY, maxX, maxY }
}
