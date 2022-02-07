import { LineChartSeries } from '../types'

/**
 * Default value for line color in case if we didn't get color for line from content config.
 */
export const DEFAULT_LINE_STROKE = 'var(--gray-07)'

export function getLineColor(series: LineChartSeries<any>): string {
    return series.color ?? DEFAULT_LINE_STROKE
}
