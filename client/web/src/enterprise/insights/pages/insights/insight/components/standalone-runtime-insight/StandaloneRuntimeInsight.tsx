import React, { useContext, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ParentSize } from '../../../../../../../charts'
import {
    CategoricalBasedChartTypes,
    CategoricalChart,
    InsightCard,
    InsightCardHeader,
    InsightCardLegend,
    InsightCardLoading,
    SeriesBasedChartTypes,
    SeriesChart,
} from '../../../../../components'
import { useInsightData } from '../../../../../components/insights-view-grid/hooks/use-insight-data'
import { CodeInsightsBackendContext, LangStatsInsight, SearchRuntimeBasedInsight } from '../../../../../core'
import { InsightContentType } from '../../../../../core/types/insight/common'
import { LazyQueryStatus } from '../../../../../hooks/use-parallel-requests/use-parallel-request'
import { getTrackingTypeByInsightType, useCodeInsightViewPings } from '../../../../../pings'
import { StandaloneInsightContextMenu } from '../context-menu/StandaloneInsightContextMenu'

import styles from './StandaloneRuntimeInsight.module.scss'

interface StandaloneRuntimeInsightProps extends TelemetryProps {
    insight: SearchRuntimeBasedInsight | LangStatsInsight
    className?: string
}

export function StandaloneRuntimeInsight(props: StandaloneRuntimeInsightProps): React.ReactElement {
    const { insight, telemetryService, className } = props
    const { getBuiltInInsightData } = useContext(CodeInsightsBackendContext)

    const insightCardReference = useRef<HTMLDivElement>(null)

    const { state, isVisible } = useInsightData(
        useMemo(() => () => getBuiltInInsightData({ insight }), [getBuiltInInsightData, insight]),
        insightCardReference
    )

    // Visual line chart settings
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)

    const { trackDatumClicks, trackMouseLeave, trackMouseEnter } = useCodeInsightViewPings({
        telemetryService,
        insightType: getTrackingTypeByInsightType(insight.type),
    })

    return (
        <InsightCard
            ref={insightCardReference}
            data-testid={`insight-card.${insight.id}`}
            className={classNames(styles.chart, className)}
            onMouseEnter={trackMouseEnter}
            onMouseLeave={trackMouseLeave}
        >
            <InsightCardHeader title={insight.title}>
                <StandaloneInsightContextMenu
                    insight={insight}
                    zeroYAxisMin={zeroYAxisMin}
                    onToggleZeroYAxisMin={setZeroYAxisMin}
                />
            </InsightCardHeader>
            {state.status === LazyQueryStatus.Loading || !isVisible ? (
                <InsightCardLoading>Loading code insight</InsightCardLoading>
            ) : state.status === LazyQueryStatus.Error ? (
                <ErrorAlert error={state.error} />
            ) : (
                <>
                    <ParentSize className={styles.chartParentSize}>
                        {parent =>
                            state.data.type === InsightContentType.Series ? (
                                <SeriesChart
                                    type={SeriesBasedChartTypes.Line}
                                    width={parent.width}
                                    height={parent.height}
                                    zeroYAxisMin={zeroYAxisMin}
                                    locked={insight.isFrozen}
                                    onDatumClick={trackDatumClicks}
                                    {...state.data.content}
                                />
                            ) : (
                                <CategoricalChart
                                    type={CategoricalBasedChartTypes.Pie}
                                    width={parent.width}
                                    height={parent.height}
                                    locked={insight.isFrozen}
                                    onDatumLinkClick={trackDatumClicks}
                                    {...state.data.content}
                                />
                            )
                        }
                    </ParentSize>
                    {state.data.type === InsightContentType.Series && (
                        <InsightCardLegend series={state.data.content.series} className="mt-3" />
                    )}
                </>
            )}
        </InsightCard>
    )
}
