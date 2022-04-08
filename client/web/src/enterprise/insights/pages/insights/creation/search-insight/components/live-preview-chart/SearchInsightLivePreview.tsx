import React, { useContext, useMemo } from 'react'

import RefreshIcon from 'mdi-react/RefreshIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, useDeepMemo } from '@sourcegraph/wildcard'

import { getLineColor, LegendItem, LegendList, ParentSize } from '../../../../../../../../charts'
import { getSanitizedRepositories, useLivePreview, StateStatus } from '../../../../../../components/creation-ui-kit'
import {
    InsightCard,
    InsightCardBanner,
    InsightCardLoading,
    SeriesBasedChartTypes,
    SeriesChart,
} from '../../../../../../components/views'
import { CodeInsightsBackendContext } from '../../../../../../core/backend/code-insights-backend-context'
import { SearchBasedInsightSeries } from '../../../../../../core/types'
import { EditableDataSeries, InsightStep } from '../../types'
import { getSanitizedLine } from '../../utils/insight-sanitizer'

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
                                <SeriesChart
                                    type={SeriesBasedChartTypes.Line}
                                    width={parent.width}
                                    height={parent.height}
                                    data-testid="code-search-insight-live-preview"
                                    {...state.data}
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

                {state.status === StateStatus.Data && (
                    <LegendList className="mt-3">
                        {state.data.series.map(series => (
                            <LegendItem key={series.dataKey} color={getLineColor(series)} name={series.name} />
                        ))}
                    </LegendList>
                )}
            </InsightCard>
            {isAllReposMode && <p>Previews are only displayed if you individually list up to 50 repositories.</p>}
        </aside>
    )
}
