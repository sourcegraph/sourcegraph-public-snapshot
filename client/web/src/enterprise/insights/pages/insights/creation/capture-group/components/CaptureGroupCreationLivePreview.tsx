import React, { useContext, useMemo } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useDeepMemo } from '@sourcegraph/wildcard'

import { SeriesBasedChartTypes, SeriesChart } from '../../../../../components'
import {
    getSanitizedRepositories,
    useLivePreview,
    StateStatus,
    LivePreviewUpdateButton,
    LivePreviewCard,
    LivePreviewLoading,
    LivePreviewChart,
    LivePreviewBlurBackdrop,
    LivePreviewBanner,
    LivePreviewLegend,
    SERIES_MOCK_CHART,
} from '../../../../../components/creation-ui-kit'
import { CodeInsightsBackendContext, SeriesChartContent } from '../../../../../core'
import { InsightStep } from '../../search-insight'
import { getSanitizedCaptureQuery } from '../utils/capture-group-insight-sanitizer'

interface CaptureGroupCreationLivePreviewProps {
    disabled: boolean
    repositories: string
    query: string
    stepValue: string
    step: InsightStep
    isAllReposMode: boolean
    className?: string
}

export const CaptureGroupCreationLivePreview: React.FunctionComponent<
    React.PropsWithChildren<CaptureGroupCreationLivePreviewProps>
> = props => {
    const { disabled, repositories, query, stepValue, step, isAllReposMode, className } = props
    const { getCaptureInsightContent } = useContext(CodeInsightsBackendContext)

    const settings = useDeepMemo({
        disabled,
        query: getSanitizedCaptureQuery(query.trim()),
        repositories: getSanitizedRepositories(repositories),
        step: { [step]: stepValue },
    })

    const getLivePreviewContent = useMemo(
        () => ({
            disabled: settings.disabled,
            fetcher: () => getCaptureInsightContent(settings),
        }),
        [settings, getCaptureInsightContent]
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
