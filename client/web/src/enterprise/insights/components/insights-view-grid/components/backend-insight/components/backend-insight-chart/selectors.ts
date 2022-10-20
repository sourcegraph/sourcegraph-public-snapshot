import { Series } from '@sourcegraph/wildcard/';

import { BackendInsightData, BackendInsightSeries, InsightContentType } from '../../types';

/**
 * Even if you have a big enough width for putting legend aside
 * we should enable this mode only if line chart has more than N series
 */
const MINIMAL_SERIES_FOR_ASIDE_LEGEND = 3

export const isManyKeysInsight = (data: BackendInsightData): boolean => {
    if (data.type === InsightContentType.Series) {
        return data.series.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
    }

    return data.content.data.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
}

export const hasSeriesError = (series: BackendInsightSeries): boolean =>
    series.status.errors.length > 0

export const hasNoData = (data: BackendInsightData): boolean => {
    if (data.type === InsightContentType.Series) {
        return data.series.every(series => series.data.length === 0)
    }

    // If all datum have zero matches render no data layout. We need to
    // handle it explicitly on the frontend since backend returns manually
    // defined series with empty points in case of no matches for generated
    // series.
    return data.content.data.every(datum => datum.value === 0)
}

export function getLineColor(series: Series<any>): string {
    return series.color ?? 'var(--gray-07)'
}
