import React, { useContext, useMemo } from 'react'

import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useDeepMemo } from '@sourcegraph/wildcard'

import { LegendItem, LegendList, ParentSize } from '../../../../../../../../charts'
import { getSanitizedRepositories, useLivePreview, StateStatus } from '../../../../../../components/creation-ui-kit'
import {
    InsightCard,
    InsightCardBanner,
    InsightCardHeader,
    InsightCardLoading,
    SeriesBasedChartTypes,
    SeriesChart,
} from '../../../../../../components/views'
import { CodeInsightsBackendContext } from '../../../../../../core/backend/code-insights-backend-context'
import { CodeInsightTrackType, useCodeInsightViewPings } from '../../../../../../pings'
import {
    DATA_SERIES_COLORS,
    DEFAULT_MOCK_CHART_CONTENT,
    EditableDataSeries,
} from '../../../../../insights/creation/search-insight'

import styles from './DynamicInsightPreview.module.scss'

const createExampleDataSeries = (query: string): EditableDataSeries[] => [
    {
        query,
        valid: true,
        edit: false,
        id: '1',
        name: 'TODOs',
        stroke: DATA_SERIES_COLORS.ORANGE,
    },
]

interface DynamicInsightPreviewProps extends TelemetryProps {
    disabled: boolean
    repositories: string
    query: string
    className?: string
}

export const DynamicInsightPreview: React.FunctionComponent<DynamicInsightPreviewProps> = props => {
    const { disabled, repositories, query, className, telemetryService } = props

    const { getSearchInsightContent } = useContext(CodeInsightsBackendContext)

    // Compare live insight settings with deep check to avoid unnecessary
    // search insight content fetching
    const settings = useDeepMemo({
        series: createExampleDataSeries(query),
        repositories: getSanitizedRepositories(repositories),
        step: { months: 2 },
        disabled,
    })

    const getLivePreviewContent = useMemo(
        () => ({
            disabled: settings.disabled,
            fetcher: () => getSearchInsightContent(settings),
        }),
        [settings, getSearchInsightContent]
    )

    const { state } = useLivePreview(getLivePreviewContent)

    const { trackMouseEnter, trackMouseLeave, trackDatumClicks } = useCodeInsightViewPings({
        telemetryService,
        insightType: CodeInsightTrackType.InProductLandingPageInsight,
    })

    return (
        <InsightCard className={classNames(className, styles.insightCard)}>
            <InsightCardHeader title="In-line TODO statements" />
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
                                {...state.data}
                            />
                        ) : (
                            <>
                                <SeriesChart
                                    type={SeriesBasedChartTypes.Line}
                                    width={parent.width}
                                    height={parent.height}
                                    className={styles.chartWithMock}
                                    onMouseEnter={trackMouseEnter}
                                    onMouseLeave={trackMouseLeave}
                                    onDatumClick={trackDatumClicks}
                                    {...DEFAULT_MOCK_CHART_CONTENT}
                                />
                                <InsightCardBanner className={styles.disableBanner}>
                                    The chart preview will be shown here once you have filled out the repositories and
                                    series fields.
                                </InsightCardBanner>
                            </>
                        )
                    }
                </ParentSize>
            )}
            {state.status === StateStatus.Data && (
                <LegendList className="mt-3">
                    <LegendItem color={DATA_SERIES_COLORS.ORANGE} name="TODOs" />
                </LegendList>
            )}
        </InsightCard>
    )
}
