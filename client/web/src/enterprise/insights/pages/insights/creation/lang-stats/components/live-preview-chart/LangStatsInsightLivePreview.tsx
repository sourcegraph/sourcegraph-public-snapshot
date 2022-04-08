import React, { useContext, useMemo } from 'react'

import RefreshIcon from 'mdi-react/RefreshIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, useDeepMemo } from '@sourcegraph/wildcard'

import { ParentSize } from '../../../../../../../../charts'
import { useLivePreview, StateStatus } from '../../../../../../components/creation-ui-kit'
import {
    CategoricalBasedChartTypes,
    CategoricalChart,
    InsightCard,
    InsightCardBanner,
    InsightCardLoading,
} from '../../../../../../components/views'
import { CodeInsightsBackendContext } from '../../../../../../core/backend/code-insights-backend-context'

import { DEFAULT_PREVIEW_MOCK } from './live-preview-mock-data'

import styles from './LangStatsInsightLivePreview.module.scss'

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
export const LangStatsInsightLivePreview: React.FunctionComponent<LangStatsInsightLivePreviewProps> = props => {
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
            <Button variant="icon" disabled={disabled} onClick={update}>
                Live preview <RefreshIcon size="1rem" />
            </Button>

            <InsightCard className={styles.insightCard}>
                {state.status === StateStatus.Loading ? (
                    <InsightCardLoading>Loading code insight</InsightCardLoading>
                ) : state.status === StateStatus.Error ? (
                    <ErrorAlert error={state.error} />
                ) : (
                    <ParentSize className={styles.chartBlock}>
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
                                    <CategoricalChart
                                        type={CategoricalBasedChartTypes.Pie}
                                        width={parent.width}
                                        height={parent.height}
                                        className={styles.chartWithMock}
                                        {...DEFAULT_PREVIEW_MOCK}
                                    />
                                    <InsightCardBanner className={styles.disableBanner}>
                                        The chart preview will be shown here once you have filled out the repository
                                        field.
                                    </InsightCardBanner>
                                </>
                            )
                        }
                    </ParentSize>
                )}
            </InsightCard>
        </aside>
    )
}
