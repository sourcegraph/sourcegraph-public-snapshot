import React, { useContext, useMemo } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useDeepMemo, Text } from '@sourcegraph/wildcard'

import { Series } from '../../../../../charts'
import { BarChart } from '../../../../../charts/components/bar-chart/BarChart'
import { GroupByField } from '../../../../../graphql-operations'
import {
    LivePreviewUpdateButton,
    LivePreviewCard,
    LivePreviewLoading,
    LivePreviewChart,
    LivePreviewBlurBackdrop,
    LivePreviewBanner,
    LivePreviewLegend,
    getSanitizedRepositories,
    useLivePreview,
    StateStatus,
    COMPUTE_MOCK_CHART,
} from '../../../components'
import { BackendInsightDatum, CategoricalChartContent, CodeInsightsBackendContext } from '../../../core'

interface LanguageUsageDatum {
    name: string
    value: number
    fill: string
    linkURL: string
    group?: string
}

interface ComputeLivePreviewProps {
    disabled: boolean
    repositories: string
    className?: string
    series: {
        query: string
        label: string
        stroke: string
        groupBy?: GroupByField
    }[]
}

export const ComputeLivePreview: React.FunctionComponent<ComputeLivePreviewProps> = props => {
    // For the purposes of building out this component before the backend is ready
    // we are using the standard "line series" type data.
    // TODO after backend is merged, remove update the series value to use that structure
    const { disabled, repositories, series, className } = props
    const { getInsightPreviewContent: getLivePreviewContent } = useContext(CodeInsightsBackendContext)

    const sanitizedSeries = series.map(srs => ({
        query: srs.query,
        label: srs.label,
        stroke: srs.stroke,
        groupBy: srs.groupBy,
    }))

    const settings = useDeepMemo({
        disabled,
        repositories: getSanitizedRepositories(repositories),
        series: sanitizedSeries,
        // TODO: Revisit this hardcoded value. Compute does not use it, but it's still required
        //  for `searchInsightPreview`
        step: {
            days: 1,
        },
    })

    const getLivePreview = useMemo(
        () => ({
            disabled: settings.disabled,
            fetcher: () => getLivePreviewContent(settings),
        }),
        [settings, getLivePreviewContent]
    )

    const { state, update } = useLivePreview(getLivePreview)

    return (
        <aside className={className}>
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
                                    data={mapSeriesToCompute(state.data.series)}
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
                    <LivePreviewLegend series={state.data.series as Series<unknown>[]} />
                )}
            </LivePreviewCard>

            <Text className="mt-4 pl-2">
                <strong>Timeframe:</strong> May 20, 2022 - Oct 20, 2022
            </Text>
        </aside>
    )
}

const mapSeriesToCompute = (series: Series<BackendInsightDatum>[]): LanguageUsageDatum[] =>
    series.map(series => ({
        group: series.name,
        name: series.name,
        value: series.data[0].value ?? 0,
        fill: series.color ?? 'var(--blue)',
        linkURL: series.data[0].link ?? '',
    }))
