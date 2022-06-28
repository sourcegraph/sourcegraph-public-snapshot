import React, { useContext, useMemo } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useDeepMemo } from '@sourcegraph/wildcard'

import { CategoricalBasedChartTypes, CategoricalChart } from '../../../../../../components'
import {
    LivePreviewBanner,
    LivePreviewBlurBackdrop,
    LivePreviewCard,
    LivePreviewChart,
    LivePreviewLoading,
    LivePreviewUpdateButton,
    useLivePreview,
    StateStatus,
} from '../../../../../../components/creation-ui-kit'
import { CodeInsightsBackendContext, CategoricalChartContent } from '../../../../../../core'

import { DEFAULT_PREVIEW_MOCK } from './constants'

export interface LangStatsInsightLivePreviewProps {
    /**
     * Disable prop to disable live preview.
     * Used in a consumer of this component when some required fields
     * for live preview are invalid.
     */
    disabled?: boolean
    repository: string
    threshold: number
    className?: string
}

/**
 * Displays live preview chart for creation UI with the latest insights settings
 * from creation UI form.
 */
export const LangStatsInsightLivePreview: React.FunctionComponent<
    React.PropsWithChildren<LangStatsInsightLivePreviewProps>
> = props => {
    const { repository = '', threshold, disabled = false, className } = props
    const { getLangStatsInsightContent } = useContext(CodeInsightsBackendContext)

    const settings = useDeepMemo({
        repository: repository.trim(),
        otherThreshold: threshold / 100,
        disabled,
    })

    const getLivePreviewContent = useMemo(
        () => ({
            disabled: settings.disabled,
            fetcher: () => getLangStatsInsightContent(settings),
        }),
        [settings, getLangStatsInsightContent]
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
                                <CategoricalChart
                                    type={CategoricalBasedChartTypes.Pie}
                                    width={parent.width}
                                    height={parent.height}
                                    {...state.data}
                                />
                            ) : (
                                <>
                                    <LivePreviewBlurBackdrop
                                        as={CategoricalChart}
                                        type={CategoricalBasedChartTypes.Pie}
                                        width={parent.width}
                                        height={parent.height}
                                        // We cast to unknown here because ForwardReferenceComponent
                                        // doesn't support inferring as component with generic.
                                        {...(DEFAULT_PREVIEW_MOCK as CategoricalChartContent<unknown>)}
                                    />

                                    <LivePreviewBanner>
                                        The chart preview will be shown here once you have filled out the repository
                                        field.
                                    </LivePreviewBanner>
                                </>
                            )
                        }
                    </LivePreviewChart>
                )}
            </LivePreviewCard>
        </aside>
    )
}
