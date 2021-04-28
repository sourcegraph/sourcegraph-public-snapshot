import { ChartAxis } from 'sourcegraph'

import { Accessors } from '../types'

/**
 * Returns accessors map which allows charts to get right values from datum object
 * One accessor for x - time axis and a map for different series of data for y axis
 */
export function generateAccessors<Datum extends object>(
    xAxis: ChartAxis<keyof Datum, Datum>,
    series: { dataKey: keyof Datum }[]
): Accessors<Datum, keyof Datum> {
    const { dataKey: xDataKey, scale = 'time' } = xAxis

    return {
        x: data =>
            scale === 'time'
                ? // as unknown as string quick hack for cast Datum[keyof Datum] to string
                  // fix that when we will have a value type for LineChartContent<D> generic
                  new Date((data?.[xDataKey] as unknown) as number)
                : // In case if we got linear scale we have to operate with numbers
                  +(data?.[xDataKey] ?? 0),
        y: series.reduce((accessors, currentLine) => {
            const { dataKey } = currentLine
            // as unknown as string quick hack for cast Datum[keyof Datum] to string
            // fix that when we will have a value type for LineChartContent<D> generic
            const key = (dataKey as unknown) as keyof Datum

            accessors[key] = data => +data[dataKey]

            return accessors
        }, {} as Record<keyof Datum, (data: Datum) => any>),
    }
}
