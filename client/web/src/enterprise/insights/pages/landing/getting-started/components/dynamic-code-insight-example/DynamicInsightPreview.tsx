import React, { useContext, useMemo } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Series, useDeepMemo } from '@sourcegraph/wildcard'

import { useSeriesToggle } from '../../../../../../../insights/utils/use-series-toggle'
import {
    SeriesBasedChartTypes,
    SeriesChart,
    getSanitizedRepositories,
    useLivePreview,
    StateStatus,
    LivePreviewCard,
    LivePreviewHeader,
    LivePreviewLoading,
    LivePreviewChart,
    LivePreviewBlurBackdrop,
    LivePreviewBanner,
    LivePreviewLegend,
    SERIES_MOCK_CHART,
} from '../../../../../components'
import { DATA_SERIES_COLORS } from '../../../../../constants'
import {
    CodeInsightsBackendContext,
    SeriesChartContent,
    SeriesPreviewSettings
} from '../../../../../core'
import { CodeInsightTrackType, useCodeInsightViewPings } from '../../../../../pings'

const createExampleDataSeries = (query: string): SeriesPreviewSettings[] => [
    {
        query,
        label: 'TODOs',
        stroke: DATA_SERIES_COLORS.ORANGE,
    },
]

interface DynamicInsightPreviewProps extends TelemetryProps {
    disabled: boolean
    repositories: string
    query: string
    className?: string
}

export const DynamicInsightPreview: React.FunctionComponent<
    React.PropsWithChildren<DynamicInsightPreviewProps>
> = props => {
    const { disabled, repositories, query, className, telemetryService } = props

    const { getInsightPreviewContent } = useContext(CodeInsightsBackendContext)
    const seriesToggleState = useSeriesToggle()

    // Compare live insight settings with deep check to avoid unnecessary
    // search insight content fetching
    const settings = useDeepMemo({
        series: createExampleDataSeries(query),
        repositories: getSanitizedRepositories(repositories),
        step: { months: 2 },
        disabled,
    })

    const getLivePreviewContent = useMemo(
        () => ({
            disabled: settings.disabled,
            fetcher: () => getInsightPreviewContent(settings),
        }),
        [settings, getInsightPreviewContent]
    )

    const { state } = useLivePreview(getLivePreviewContent)

    const { trackMouseEnter, trackMouseLeave, trackDatumClicks } = useCodeInsightViewPings({
        telemetryService,
        insightType: CodeInsightTrackType.InProductLandingPageInsight,
    })

    return (
        <LivePreviewCard className={className}>
            <LivePreviewHeader title="In-line TODO statements" />
            {state.status === StateStatus.Loading ? (
                <LivePreviewLoading>Loading code insight</LivePreviewLoading>
            ) : state.status === StateStatus.Error ? (
                <ErrorAlert error={state.error} />
            ) : (
                <LivePreviewChart>
                    {parent =>
                        state.status === StateStatus.Data ? (
                            <SeriesChart
                                type={SeriesBasedChartTypes.Line}
                                width={parent.width}
                                height={parent.height}
                                seriesToggleState={seriesToggleState}
                                {...state.data}
                            />
                        ) : (
                            <>
                                <LivePreviewBlurBackdrop
                                    as={SeriesChart}
                                    type={SeriesBasedChartTypes.Line}
                                    width={parent.width}
                                    height={parent.height}
                                    onMouseEnter={trackMouseEnter}
                                    onMouseLeave={trackMouseLeave}
                                    onDatumClick={trackDatumClicks}
                                    seriesToggleState={seriesToggleState}
                                    // We cast to unknown here because ForwardReferenceComponent
                                    // doesn't support inferring as component with generic.
                                    {...(SERIES_MOCK_CHART as SeriesChartContent<unknown>)}
                                />
                                <LivePreviewBanner>
                                    The chart preview will be shown here once you have filled out the repositories and
                                    series fields.
                                </LivePreviewBanner>
                            </>
                        )
                    }
                </LivePreviewChart>
            )}

            {state.status === StateStatus.Data && <LivePreviewLegend series={state.data.series as Series<unknown>[]} />}
        </LivePreviewCard>
    )
}
