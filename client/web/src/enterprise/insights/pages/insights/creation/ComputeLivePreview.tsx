import React, { useContext, useMemo } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useDeepMemo, Text } from '@sourcegraph/wildcard'

import { BarChart } from '../../../../../charts/components/bar-chart/BarChart'
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
    SERIES_MOCK_CHART,
} from '../../../components'
import { CodeInsightsBackendContext } from '../../../core'

import { getSanitizedCaptureQuery } from './capture-group/utils/capture-group-insight-sanitizer'
import { InsightStep } from './search-insight'

interface LineChartLivePreviewProps {
    disabled: boolean
    repositories: string
    stepValue: string
    step: InsightStep
    className?: string
    series: {
        query: string
        label: string
        generatedFromCaptureGroup: boolean
        stroke: string
    }[]
}

export const ComputeLivePreview: React.FunctionComponent<
    React.PropsWithChildren<LineChartLivePreviewProps>
> = props => {
    const { disabled, repositories, stepValue, step, series, className } = props
    const { getInsightPreviewContent: getLivePreviewContent } = useContext(CodeInsightsBackendContext)

    const sanitizedSeries = series.map(srs => {
        const sanitizer = srs.generatedFromCaptureGroup ? getSanitizedCaptureQuery : (query: string) => query
        return {
            query: sanitizer(srs.query),
            generatedFromCaptureGroup: srs.generatedFromCaptureGroup,
            label: srs.label,
            stroke: srs.stroke,
        }
    })

    const settings = useDeepMemo({
        disabled,
        repositories: getSanitizedRepositories(repositories),
        step: { [step]: stepValue },
        series: sanitizedSeries,
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
                                    data-testid="code-search-insight-live-preview"
                                    data={state.data.series}
                                    getDatumName={(datum: any) => datum.name}
                                    getDatumValue={(datum: any) => datum.value}
                                    getDatumColor={(datum: any) => datum.color}
                                />
                            ) : (
                                <>
                                    <LivePreviewBlurBackdrop
                                        as={BarChart}
                                        width={parent.width}
                                        height={parent.height}
                                        data={SERIES_MOCK_CHART.series}
                                        getDatumName={(datum: any) => datum.name}
                                        getDatumValue={(datum: any) => datum.value}
                                        getDatumColor={(datum: any) => datum.color}
                                    />
                                    <LivePreviewBanner>You’ll see your insight’s chart preview here</LivePreviewBanner>
                                </>
                            )
                        }
                    </LivePreviewChart>
                )}

                {state.status === StateStatus.Data && <LivePreviewLegend series={state.data.series} />}
            </LivePreviewCard>

            <Text className="mt-4 pl-2">
                <b>Timeframe:</b> May 20, 2022 - Oct 20, 2022
            </Text>
        </aside>
    )
}
