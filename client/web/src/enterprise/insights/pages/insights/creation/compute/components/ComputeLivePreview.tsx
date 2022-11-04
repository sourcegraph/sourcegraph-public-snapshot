import { HTMLAttributes, useContext, useMemo, FC } from 'react'

import { groupBy } from 'lodash'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useDeepMemo, BarChart, LegendList, LegendItem, Series } from '@sourcegraph/wildcard'

import { GroupByField } from '../../../../../../../graphql-operations'
import {
    LivePreviewUpdateButton,
    LivePreviewCard,
    LivePreviewLoading,
    LivePreviewChart,
    LivePreviewBlurBackdrop,
    LivePreviewBanner,
    getSanitizedRepositories,
    useLivePreview,
    StateStatus,
    COMPUTE_MOCK_CHART,
    EditableDataSeries,
} from '../../../../../components'
import {
    BackendInsightDatum,
    CategoricalChartContent,
    CodeInsightsBackendContext,
    SeriesPreviewSettings,
} from '../../../../../core'

interface LanguageUsageDatum {
    name: string
    value: number
    fill: string
    group?: string
}

interface ComputeLivePreviewProps extends HTMLAttributes<HTMLElement> {
    disabled: boolean
    repositories: string
    className?: string
    groupBy: GroupByField
    series: EditableDataSeries[]
}

export const ComputeLivePreview: FC<ComputeLivePreviewProps> = props => {
    const { disabled, repositories, series, groupBy, ...attribute } = props
    const { getInsightPreviewContent } = useContext(CodeInsightsBackendContext)

    const settings = useDeepMemo({
        disabled,
        repositories: getSanitizedRepositories(repositories),
        // For the purposes of building out this component before the backend is ready
        // we are using the standard "line series" type data.
        // TODO after backend is merged, remove update the series value to use that structure
        series: series.map<SeriesPreviewSettings>(srs => ({
            query: srs.query,
            label: srs.name,
            stroke: srs.stroke ?? 'blue',
            generatedFromCaptureGroup: true,
            groupBy,
        })),
        // TODO: Revisit this hardcoded value. Compute does not use it, but it's still required
        //  for `searchInsightPreview`
        step: { days: 1 },
    })

    const getLivePreview = useMemo(
        () => ({
            disabled: settings.disabled,
            fetcher: () => getInsightPreviewContent(settings),
        }),
        [settings, getInsightPreviewContent]
    )

    const { state, update } = useLivePreview(getLivePreview)

    return (
        <aside {...attribute}>
            <LivePreviewUpdateButton disabled={disabled} onClick={update} />

            <LivePreviewCard>
                {state.status === StateStatus.Loading ? (
                    <LivePreviewLoading>Loading code insight</LivePreviewLoading>
                ) : state.status === StateStatus.Error ? (
                    <ErrorAlert error={state.error} />
                ) : (
                    <LivePreviewChart>
                        {parent =>
                            state.status === StateStatus.Data ? (
                                <BarChart
                                    width={parent.width}
                                    height={parent.height}
                                    data={mapSeriesToCompute(settings.series, state.data.series)}
                                    getCategory={(datum: LanguageUsageDatum) => datum.group}
                                    getDatumName={(datum: LanguageUsageDatum) => datum.name}
                                    getDatumValue={(datum: LanguageUsageDatum) => datum.value}
                                    getDatumColor={(datum: LanguageUsageDatum) => datum.fill}
                                />
                            ) : (
                                <>
                                    <LivePreviewBlurBackdrop
                                        as={BarChart}
                                        width={parent.width}
                                        height={parent.height}
                                        getCategory={(datum: unknown) => (datum as LanguageUsageDatum).group}
                                        // We cast to unknown here because ForwardReferenceComponent
                                        // doesn't support types inferring if component has a generic parameter.
                                        {...(COMPUTE_MOCK_CHART as CategoricalChartContent<unknown>)}
                                    />
                                    <LivePreviewBanner>You’ll see your insight’s chart preview here</LivePreviewBanner>
                                </>
                            )
                        }
                    </LivePreviewChart>
                )}

                {state.status === StateStatus.Data && (
                    <LegendList className="mt-3">
                        {settings.series.map(series => (
                            <LegendItem key={series.label} color={series.stroke} name={series.label} />
                        ))}
                    </LegendList>
                )}
            </LivePreviewCard>
        </aside>
    )
}

const mapSeriesToCompute = (
    seriesDefinitions: SeriesPreviewSettings[],
    series: Series<BackendInsightDatum>[]
): LanguageUsageDatum[] => {
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
const getComputeSeriesColor = (series: SeriesPreviewSettings): string => series?.stroke ?? 'var(--blue)'
