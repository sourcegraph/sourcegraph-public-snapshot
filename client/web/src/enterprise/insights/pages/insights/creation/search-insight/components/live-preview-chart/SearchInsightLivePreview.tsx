import React from 'react'

import RefreshIcon from 'mdi-react/RefreshIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { Button, useDeepMemo } from '@sourcegraph/wildcard'

import { getLineColor, LegendItem, LegendList, ParentSize } from '../../../../../../../../charts'
import { getSanitizedRepositories } from '../../../../../../components/creation-ui-kit'
import {
    InsightCard,
    InsightCardBanner,
    InsightCardLoading,
    SeriesBasedChartTypes,
    SeriesChart,
} from '../../../../../../components/views'
import { SearchBasedInsightSeries } from '../../../../../../core/types'
import { EditableDataSeries, InsightStep } from '../../types'
import { getSanitizedLine } from '../../utils/insight-sanitizer'

import { useSearchBasedLivePreviewContent } from './hooks/use-search-based-live-preview-content'
import { DEFAULT_MOCK_CHART_CONTENT } from './live-preview-mock-data'

import styles from './SearchInsightLivePreview.module.scss'

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
export const SearchInsightLivePreview: React.FunctionComponent<SearchInsightLivePreviewProps> = props => {
    const { series, repositories, step, stepValue, disabled = false, isAllReposMode, className } = props

    // Compare live insight settings with deep check to avoid unnecessary
    // search insight content fetching
    const previewSetting = useDeepMemo({
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

    const { loading, dataOrError, update } = useSearchBasedLivePreviewContent(previewSetting)

    return (
        <aside className={className}>
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
                                <SeriesChart
                                    type={SeriesBasedChartTypes.Line}
                                    width={parent.width}
                                    height={parent.height}
                                    data-testid="code-search-insight-live-preview"
                                    {...dataOrError}
                                />
                            ) : (
                                <>
                                    <SeriesChart
                                        type={SeriesBasedChartTypes.Line}
                                        width={parent.width}
                                        height={parent.height}
                                        className={styles.chartWithMock}
                                        {...DEFAULT_MOCK_CHART_CONTENT}
                                    />
                                    <InsightCardBanner className={styles.disableBanner}>
                                        {isAllReposMode
                                            ? 'Live previews are currently not available for insights running over all repositories.'
                                            : 'The chart preview will be shown here once you have filled out the repositories and series fields.'}
                                    </InsightCardBanner>
                                </>
                            )
                        }
                    </ParentSize>
                )}

                {dataOrError && !isErrorLike(dataOrError) && (
                    <LegendList className="mt-3">
                        {dataOrError.series.map(series => (
                            <LegendItem key={series.dataKey} color={getLineColor(series)} name={series.name} />
                        ))}
                    </LegendList>
                )}
            </InsightCard>
            {isAllReposMode && <p>Previews are only displayed if you individually list up to 50 repositories.</p>}
        </aside>
    )
}
