import React, { useContext, useMemo } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useDeepMemo } from '@sourcegraph/wildcard'

import { SeriesBasedChartTypes, SeriesChart } from '../../../../../components'
import {
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
} from '../../../../../components/creation-ui-kit'
import { CodeInsightsBackendContext, SeriesChartContent } from '../../../../../core'
import { CodeInsightTrackType, useCodeInsightViewPings } from '../../../../../pings'
import { DATA_SERIES_COLORS, EditableDataSeries } from '../../../../insights/creation/search-insight'

const createExampleDataSeries = (query: string): EditableDataSeries[] => [
    {
        query,
        valid: true,
        edit: false,
        id: '1',
        name: 'TODOs',
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

    const { getSearchInsightContent } = useContext(CodeInsightsBackendContext)

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
            fetcher: () => getSearchInsightContent(settings),
        }),
        [settings, getSearchInsightContent]
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

            {state.status === StateStatus.Data && <LivePreviewLegend series={state.data.series} />}
        </LivePreviewCard>
    )
}
