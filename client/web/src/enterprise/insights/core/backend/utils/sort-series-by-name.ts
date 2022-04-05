import semver from 'semver'
import { LineChartSeries } from 'sourcegraph'

import { SeriesDataset } from './create-line-chart-content'

export type SortSeriesByNameParameter = Pick<LineChartSeries<SeriesDataset>, 'name'>

export function sortSeriesByName(seriesA: SortSeriesByNameParameter, seriesB: SortSeriesByNameParameter): number {
    if (!seriesA.name || !seriesB.name) {
        return 0
    }

    if (semver.valid(seriesA.name) && semver.valid(seriesB.name)) {
        return semver.gt(seriesA.name, seriesB.name) ? 1 : -1
    }
    return seriesA.name.localeCompare(seriesB.name) || 0
}
