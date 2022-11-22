import { FC, HTMLAttributes } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useDeepMemo, Text, Series, useDebounce } from '@sourcegraph/wildcard'

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
    getSanitizedRepositories,
    SERIES_MOCK_CHART,
} from '../../../components'
import { useLivePreviewSeriesInsight, LivePreviewStatus } from '../../../core/hooks/live-preview-insight'

import { getSanitizedCaptureQuery } from './capture-group/utils/capture-group-insight-sanitizer'
import { InsightStep } from './search-insight'

export interface LivePreviewSeries {
    query: string
    label: string
    generatedFromCaptureGroup: boolean
    stroke: string
}

interface LineChartLivePreviewProps extends HTMLAttributes<HTMLElement> {
    disabled: boolean
    repositories: string
    stepValue: string
    step: InsightStep
    isAllReposMode: boolean
    series: LivePreviewSeries[]
}

export const LineChartLivePreview: FC<LineChartLivePreviewProps> = props => {
    const { disabled, repositories, stepValue, step, series, isAllReposMode, ...attributes } = props
    const seriesToggleState = useSeriesToggle()

    const settings = useDebounce(
        useDeepMemo({
            repositories: getSanitizedRepositories(repositories),
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
        skip: disabled,
        ...settings,
    })

    return (
        <aside {...attributes}>
            <LivePreviewUpdateButton disabled={disabled} onClick={refetch} />

            <LivePreviewCard>
                {state.status === LivePreviewStatus.Loading ? (
                    <LivePreviewLoading>Loading code insight</LivePreviewLoading>
                ) : state.status === LivePreviewStatus.Error ? (
                    <ErrorAlert error={state.error} />
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
                                        {isAllReposMode
                                            ? 'Live previews are currently not available for insights running over all repositories.'
                                            : 'The chart preview will be shown here once you have filled out the repositories and series fields.'}
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

            {isAllReposMode && (
                <Text className="mt-2">
                    Previews are only displayed if you individually list up to 50 repositories.
                </Text>
            )}
        </aside>
    )
}
