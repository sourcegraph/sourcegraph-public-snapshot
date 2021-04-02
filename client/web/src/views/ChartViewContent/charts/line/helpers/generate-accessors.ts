import { ChartAxis } from 'sourcegraph'
import { Accessors } from '../types'

export function generateAccessors<Datum extends object>(
    xAxis: ChartAxis<keyof Datum, Datum>,
    series: { dataKey: keyof Datum}[]
): Accessors<Datum, string> {
    const { dataKey: xDataKey, scale = 'time' } = xAxis;

    return {
        x: data => scale === 'time'
            // as unknown as string quick hack for cast Datum[keyof Datum] to string
            // fix that when we will have a value type for LineChartContent<D> generic
            ? new Date(data?.[xDataKey] as unknown as string)
            // In case if we got linear scale we have to operate with numbers
            : +(data?.[xDataKey] ?? 0),
        y: series.reduce<Record<string, (data: Datum) => any>>((accessors, currentLine) => {
            const { dataKey } = currentLine;
            // as unknown as string quick hack for cast Datum[keyof Datum] to string
            // fix that when we will have a value type for LineChartContent<D> generic
            const key = dataKey as unknown as string;

            accessors[key] = data => +data[dataKey];

            return accessors;
        }, {})
    };
}
