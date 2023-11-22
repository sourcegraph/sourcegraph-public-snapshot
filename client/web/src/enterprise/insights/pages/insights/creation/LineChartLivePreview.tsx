import type { FC, HTMLAttributes } from 'react'

import { useDeepMemo, type Series, useDebounce, ErrorAlert } from '@sourcegraph/wildcard'

import { useSeriesToggle } from '../../../../../insights/utils/use-series-toggle'
import {
    SeriesChart,
    SeriesBasedChartTypes,
    LivePreviewUpdateButton,
    LivePreviewCard,
    LivePreviewLoading,
    LivePreviewChart,
    LivePreviewBlurBackdrop,
    LivePreviewBanner,
    LivePreviewLegend,
    getSanitizedRepositoryScope,
    SERIES_MOCK_CHART,
} from '../../../components'
import { useLivePreviewSeriesInsight, LivePreviewStatus } from '../../../core'

import { getSanitizedCaptureQuery } from './capture-group/utils/capture-group-insight-sanitizer'
import type { InsightStep } from './search-insight'

export interface LivePreviewSeries {
    query: string
    label: string
    generatedFromCaptureGroup: boolean
    stroke: string
}

interface LineChartLivePreviewProps extends HTMLAttributes<HTMLElement> {
    disabled: boolean
    repositories: string[]
    repoQuery: string | undefined
    repoMode: string
    stepValue: string
    step: InsightStep
    series: LivePreviewSeries[]
}

export const LineChartLivePreview: FC<LineChartLivePreviewProps> = props => {
    const { disabled, repositories, repoQuery, repoMode, stepValue, step, series, ...attributes } = props
    const seriesToggleState = useSeriesToggle()

    const settings = useDebounce(
        useDeepMemo({
            disabled,
            repoScope: getSanitizedRepositoryScope(repositories, repoQuery, repoMode),
            step: { [step]: stepValue },
            series: series.map(srs => {
                const sanitizer = srs.generatedFromCaptureGroup
                    ? getSanitizedCaptureQuery
                    : (query: string) => query.trim()

                return {
                    query: sanitizer(srs.query),
                    generatedFromCaptureGroups: srs.generatedFromCaptureGroup,
                    label: srs.label,
                    stroke: srs.stroke,
                }
            }),
        }),
        500
    )

    const { state, refetch } = useLivePreviewSeriesInsight({
        // If disabled goes from true to false then cancel live preview series fetching
        // immediately, when it goes from false to true wait a little *use debounced
        // value only when run preview search
        skip: disabled || settings.disabled,
        ...settings,
    })

    return (
        <aside {...attributes}>
            <LivePreviewUpdateButton disabled={disabled} onClick={refetch} />

            <LivePreviewCard className="flex-1">
                {state.status === LivePreviewStatus.Loading ? (
                    <LivePreviewLoading>Loading code insight</LivePreviewLoading>
                ) : state.status === LivePreviewStatus.Error ? (
                    <ErrorAlert error={state.error} className="m-0" />
                ) : (
                    <LivePreviewChart>
                        {parent =>
                            state.status === LivePreviewStatus.Data ? (
                                <SeriesChart
                                    type={SeriesBasedChartTypes.Line}
                                    width={parent.width}
                                    height={parent.height}
                                    data-testid="code-search-insight-live-preview"
                                    seriesToggleState={seriesToggleState}
                                    series={state.data}
                                />
                            ) : (
                                <>
                                    <LivePreviewBlurBackdrop
                                        as={SeriesChart}
                                        type={SeriesBasedChartTypes.Line}
                                        width={parent.width}
                                        height={parent.height}
                                        // We cast to unknown here because ForwardReferenceComponent
                                        // doesn't support inferring as component with generic.
                                        series={SERIES_MOCK_CHART as Series<unknown>[]}
                                    />
                                    <LivePreviewBanner>
                                        The chart preview will be shown here once you have filled out the repositories
                                        and series fields.
                                    </LivePreviewBanner>
                                </>
                            )
                        }
                    </LivePreviewChart>
                )}

                {state.status === LivePreviewStatus.Data && (
                    <LivePreviewLegend series={state.data as Series<unknown>[]} />
                )}
            </LivePreviewCard>
        </aside>
    )
}
