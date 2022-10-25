import { DEFAULT_FALLBACK_COLOR } from '../../../constants'
import { Series } from '../../../types'

export function getLineColor(series: Series<any>): string {
    return series.color ?? DEFAULT_FALLBACK_COLOR
}
