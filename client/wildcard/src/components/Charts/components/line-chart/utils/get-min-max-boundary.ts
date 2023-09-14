import { getDatumValue, isDatumWithValidNumber, type SeriesWithData } from './data-series-processing'

interface MinMaxBoundariesInput<D> {
    dataSeries: SeriesWithData<D>[]
    zeroYAxisMin: boolean
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
    const { dataSeries, zeroYAxisMin } = props

    let minX
    let maxX
    let minY
    let maxY

    for (const line of dataSeries) {
        for (const data of line.data) {
            minX = Math.min(+data.x, minX ?? +data.x)
            maxX = Math.max(+data.x, maxX ?? +data.x)

            if (isDatumWithValidNumber(data)) {
                minY = Math.min(getDatumValue(data), minY ?? getDatumValue(data))
                maxY = Math.max(getDatumValue(data), maxY ?? getDatumValue(data))
            }
        }
    }

    if (zeroYAxisMin) {
        minY = 0
    }

    ;[minY, maxY, minX, maxX] = [minY ?? 0, maxY ?? 0, minX ?? 0, maxX ?? 0]

    // Expand range for better ticks looking in case if we got a flat data series dataset
    ;[minY, maxY] = minY === maxY ? [maxY - maxY / 2, maxY + maxY / 2] : [minY, maxY]

    return { minX, minY, maxX, maxY }
}
