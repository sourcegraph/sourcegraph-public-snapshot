import { DEFAULT_FALLBACK_COLOR } from '../../../constants'
import { LineChartSeries } from '../types'

export function getLineColor(series: LineChartSeries<any>): string {
    return series.color ?? DEFAULT_FALLBACK_COLOR
}
