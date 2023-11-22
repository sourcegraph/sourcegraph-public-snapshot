import type { FC, HTMLAttributes } from 'react'

import { useDebounce, useDeepMemo, ErrorAlert } from '@sourcegraph/wildcard'

import {
    CategoricalBasedChartTypes,
    CategoricalChart,
    LivePreviewBanner,
    LivePreviewBlurBackdrop,
    LivePreviewCard,
    LivePreviewChart,
    LivePreviewLoading,
    LivePreviewUpdateButton,
} from '../../../../../../components'
import type { CategoricalChartContent } from '../../../../../../core'
import { LivePreviewStatus, useLivePreviewLangStatsInsight } from '../../../../../../core/hooks/live-preview-insight'

import { DEFAULT_PREVIEW_MOCK } from './constants'

export interface LangStatsInsightLivePreviewProps extends HTMLAttributes<HTMLElement> {
    /**
     * Disable prop to disable live preview.
     * Used in a consumer of this component when some required fields
     * for live preview are invalid.
     */
    disabled?: boolean
    repository: string
    threshold: number
}

/**
 * Displays live preview chart for creation UI with the latest insights settings
 * from creation UI form.
 */
export const LangStatsInsightLivePreview: FC<LangStatsInsightLivePreviewProps> = props => {
    const { repository = '', threshold, disabled = false, ...attributes } = props

    const settings = useDeepMemo({
        repository: repository.trim(),
        otherThreshold: threshold / 100,
        skip: disabled,
    })

    const debouncedSettings = useDebounce(settings, 500)
    const { state, refetch } = useLivePreviewLangStatsInsight(debouncedSettings)

    return (
        <aside {...attributes}>
            <LivePreviewUpdateButton disabled={disabled} onClick={refetch} />

            <LivePreviewCard>
                {state.status === LivePreviewStatus.Loading ? (
                    <LivePreviewLoading>Loading code insight</LivePreviewLoading>
                ) : state.status === LivePreviewStatus.Error ? (
                    <ErrorAlert error={state.error} className="m-0" />
                ) : (
                    <LivePreviewChart>
                        {parent =>
                            state.status === LivePreviewStatus.Data ? (
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
