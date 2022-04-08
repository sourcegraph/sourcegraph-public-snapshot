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
import { InsightStep } from '../../../search-insight'
import { getSanitizedCaptureQuery } from '../../utils/capture-group-insight-sanitizer'

import { MOCK_CHART_CONTENT } from './constants'

import styles from './CaptureGroupCreationLivePreview.module.scss'

interface CaptureGroupCreationLivePreviewProps {
    disabled: boolean
    repositories: string
    query: string
    stepValue: string
    step: InsightStep
    isAllReposMode: boolean
    className?: string
}

export const CaptureGroupCreationLivePreview: React.FunctionComponent<CaptureGroupCreationLivePreviewProps> = props => {
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
                                        {...MOCK_CHART_CONTENT}
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
            {isAllReposMode && (
                <p className="mt-2">Previews are only displayed if you individually list up to 50 repositories.</p>
            )}
        </aside>
    )
}
