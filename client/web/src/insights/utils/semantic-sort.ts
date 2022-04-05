import semver from 'semver'
import { LineChartSeries } from 'sourcegraph'

interface SeriesDataset {
    dateTime: number
    [seriesKey: string]: number
}

export type SortSeriesByNameParameter = Pick<LineChartSeries<SeriesDataset>, 'name'>

export function semanticSort(stringA?: string, stringB?: string): number {
    if (!stringA || !stringB) {
        return 0
    }

    if (semver.valid(stringA) && semver.valid(stringB)) {
        return semver.gt(stringA, stringB) ? 1 : -1
    }
    return stringA.localeCompare(stringB) || 0
}
