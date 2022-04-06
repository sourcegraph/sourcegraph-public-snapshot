import React from 'react'

import classNames from 'classnames'
import RefreshIcon from 'mdi-react/RefreshIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { Button, useDeepMemo } from '@sourcegraph/wildcard'

import { ParentSize } from '../../../../../../../../charts'
import {
    CategoricalBasedChartTypes,
    CategoricalChart,
    InsightCard,
    InsightCardLoading,
    InsightCardBanner,
} from '../../../../../../components/views'

import { useLangStatsPreviewContent } from './hooks/use-lang-stats-preview-content'
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
 * Displays live preview chart for creation UI with latest insights settings
 * from creation UI form.
 */
export const LangStatsInsightLivePreview: React.FunctionComponent<LangStatsInsightLivePreviewProps> = props => {
    const { repository = '', threshold, disabled = false, className } = props

    const previewSetting = useDeepMemo({
        repository: repository.trim(),
        otherThreshold: threshold / 100,
    })

    const { loading, dataOrError, update } = useLangStatsPreviewContent({ disabled, previewSetting })

    return (
        <aside className={classNames(className)}>
            <Button variant="icon" disabled={disabled} onClick={update}>
                Live preview <RefreshIcon size="1rem" />
            </Button>

            <InsightCard className={styles.insightCard}>
                {loading ? (
                    <InsightCardLoading>Loading code insight</InsightCardLoading>
                ) : isErrorLike(dataOrError) ? (
                    <ErrorAlert error={dataOrError} />
                ) : (
                    <ParentSize className={styles.chartBlock}>
                        {parent =>
                            dataOrError ? (
                                <CategoricalChart
                                    type={CategoricalBasedChartTypes.Pie}
                                    width={parent.width}
                                    height={parent.height}
                                    {...dataOrError}
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
