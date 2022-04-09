import React from 'react'

import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useDeepMemo } from '@sourcegraph/wildcard'

import { LegendItem, LegendList, ParentSize } from '../../../../../../../../charts'
import { getSanitizedRepositories } from '../../../../../../components/creation-ui-kit'
import {
    InsightCard,
    InsightCardHeader,
    InsightCardBanner,
    InsightCardLoading,
    SeriesBasedChartTypes,
    SeriesChart,
} from '../../../../../../components/views'
import { CodeInsightTrackType, useCodeInsightViewPings } from '../../../../../../pings'
import {
    DATA_SERIES_COLORS,
    DEFAULT_MOCK_CHART_CONTENT,
    EditableDataSeries,
    useSearchBasedLivePreviewContent,
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

    // Compare live insight settings with deep check to avoid unnecessary
    // search insight content fetching
    const previewSetting = useDeepMemo({
        series: createExampleDataSeries(query),
        repositories: getSanitizedRepositories(repositories),
        step: { months: 2 },
        disabled,
    })

    const { loading, dataOrError } = useSearchBasedLivePreviewContent(previewSetting)

    const { trackMouseEnter, trackMouseLeave, trackDatumClicks } = useCodeInsightViewPings({
        telemetryService,
        insightType: CodeInsightTrackType.InProductLandingPageInsight,
    })

    return (
        <InsightCard className={classNames(className, styles.insightCard)}>
            <InsightCardHeader title="In-line TODO statements" />
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
                                {...dataOrError}
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
            {dataOrError && !isErrorLike(dataOrError) && (
                <LegendList className="mt-3">
                    <LegendItem color={DATA_SERIES_COLORS.ORANGE} name="TODOs" />
                </LegendList>
            )}
        </InsightCard>
    )
}
