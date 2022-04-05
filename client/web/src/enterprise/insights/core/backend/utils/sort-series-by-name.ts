import { LineChartSeries } from 'sourcegraph'

import { SeriesDataset } from './create-line-chart-content'

export function sortSeriesByName(
    seriesA: LineChartSeries<SeriesDataset>,
    seriesB: LineChartSeries<SeriesDataset>
): number {
    return (seriesA.name && seriesB.name && seriesA.name.localeCompare(seriesB.name)) || 0
}
