import { groupBy } from 'lodash'

import { GroupByField } from '@sourcegraph/shared/src/graphql-operations'
import { Series } from '@sourcegraph/wildcard'

import { LivePreviewStatus, State } from './types'
import { Datum, SeriesWithStroke, useLivePreviewSeriesInsight } from './use-live-preview-series-insight'

interface ComputeSeries extends Omit<SeriesWithStroke, 'generatedFromCaptureGroups'> {}

export interface Props {
    skip: boolean
    repositories: string[]
    series: ComputeSeries[]
    groupBy: GroupByField
}

interface Result<R> {
    state: State<ComputeDatum[]>
    refetch: () => {}
}

export function useLivePreviewComputeInsight(props: Props): Result<ComputeDatum[]> {
    const { skip, repositories, series, groupBy } = props

    const { state, refetch } = useLivePreviewSeriesInsight({
        skip,
        repositories,
        series: series.map(srs => ({
            ...srs,
            groupBy,
            generatedFromCaptureGroups: true,
        })),
        // TODO: Revisit this hardcoded value. Compute does not use it, but it's still required
        //  for `searchInsightPreview`
        step: { days: 1 },
    })

    // Post process data from series insight preview since compute is based
    // on series live preview handler
    if (state.status === LivePreviewStatus.Data) {
        return {
            state: {
                status: LivePreviewStatus.Data,
                data: mapSeriesToCompute(series, state.data.series),
            },
            refetch,
        }
    }

    return { state, refetch }
}

export interface ComputeDatum {
    name: string
    value: number
    fill: string
    group?: string
}

const mapSeriesToCompute = (seriesDefinitions: ComputeSeries[], series: Series<Datum>[]): ComputeDatum[] => {
    const seriesGroups = groupBy(
        series.filter(series => series.name),
        series => series.name
    )

    // Group series result by seres name and sum up series value with the same name
    return Object.keys(seriesGroups).map(key =>
        seriesGroups[key].reduce(
            (memo, series) => {
                memo.value += series.data.reduce((sum, datum) => sum + (series.getYValue(datum) ?? 0), 0)

                return memo
            },
            {
                name: getComputeSeriesName(seriesGroups[key][0]),
                // We pick color only from the first series since compute-powered insight
                // can't have more than one series see https://github.com/sourcegraph/sourcegraph/issues/38832
                fill: getComputeSeriesColor(seriesDefinitions[0]),
                value: 0,
            }
        )
    )
}

const getComputeSeriesName = (series: Series<any>): string => (series.name ? series.name : 'Other')
const getComputeSeriesColor = (series: ComputeSeries): string => series?.stroke ?? 'var(--blue)'
