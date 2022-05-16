import React, { useContext, useMemo } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useDeepMemo } from '@sourcegraph/wildcard'

import { SeriesBasedChartTypes, SeriesChart } from '../../../../../components'
import {
    getSanitizedRepositories,
    useLivePreview,
    StateStatus,
    LivePreviewCard,
    LivePreviewLoading,
    LivePreviewChart,
    LivePreviewBlurBackdrop,
    LivePreviewBanner,
    LivePreviewLegend,
    LivePreviewUpdateButton,
    SERIES_MOCK_CHART,
} from '../../../../../components/creation-ui-kit'
import { CodeInsightsBackendContext, SeriesChartContent, SearchBasedInsightSeries } from '../../../../../core'
import { EditableDataSeries, InsightStep } from '../types'
import { getSanitizedLine } from '../utils/insight-sanitizer'

export interface SearchInsightLivePreviewProps {
    /**
     * Disable prop to disable live preview.
     * Used in a consumer of this component when some required fields
     * for live preview are invalid.
     */
    disabled: boolean
    repositories: string
    series: EditableDataSeries[]
    stepValue: string
    step: InsightStep
    isAllReposMode: boolean

    className?: string
}

/**
 * Displays live preview chart for creation UI with the latest insights settings
 * from creation UI form.
 */
export const SearchInsightLivePreview: React.FunctionComponent<
    React.PropsWithChildren<SearchInsightLivePreviewProps>
> = props => {
    const { series, repositories, step, stepValue, disabled = false, isAllReposMode, className } = props

    const { getSearchInsightContent } = useContext(CodeInsightsBackendContext)

    // Compare live insight settings with deep check to avoid unnecessary
    // search insight content fetching
    const settings = useDeepMemo({
        series: series
            .filter(series => series.valid)
            // Cut off all unnecessary for live preview fields in order to
            // not trigger live preview update if any of unnecessary has been updated
            // Example: edit true => false - chart shouldn't re-fetch data
            .map<SearchBasedInsightSeries>(getSanitizedLine),
        repositories: getSanitizedRepositories(repositories),
        step: { [step]: stepValue },
        disabled,
    })

    const getLivePreviewContent = useMemo(
        () => ({
            disabled: settings.disabled,
            fetcher: () => getSearchInsightContent(settings),
        }),
        [settings, getSearchInsightContent]
    )

    const { state, update } = useLivePreview(getLivePreviewContent)

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
                                <SeriesChart
                                    type={SeriesBasedChartTypes.Line}
                                    width={parent.width}
                                    height={parent.height}
                                    data-testid="code-search-insight-live-preview"
                                    {...state.data}
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

                {state.status === StateStatus.Data && <LivePreviewLegend series={state.data.series} />}
            </LivePreviewCard>
            {isAllReposMode && (
                <p className="mt-2">Previews are only displayed if you individually list up to 50 repositories.</p>
            )}
        </aside>
    )
}
