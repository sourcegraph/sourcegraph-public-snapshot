import { useContext, useMemo, FC, HTMLAttributes } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useDeepMemo, Text, Series } from '@sourcegraph/wildcard'

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
    useLivePreview,
    StateStatus,
    SERIES_MOCK_CHART,
} from '../../../components'
import { CodeInsightsBackendContext, SeriesChartContent } from '../../../core'

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
    const { getInsightPreviewContent: getLivePreviewContent } = useContext(CodeInsightsBackendContext)
    const seriesToggleState = useSeriesToggle()

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
        <aside {...attributes}>
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
                                <SeriesChart
                                    type={SeriesBasedChartTypes.Line}
                                    width={parent.width}
                                    height={parent.height}
                                    data-testid="code-search-insight-live-preview"
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
                                        seriesToggleState={seriesToggleState}
                                        // We cast to unknown here because ForwardReferenceComponent
                                        // doesn't support inferring as component with generic.
                                        {...(SERIES_MOCK_CHART as SeriesChartContent<unknown>)}
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

                {state.status === StateStatus.Data && (
                    <LivePreviewLegend series={state.data.series as Series<unknown>[]} />
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
