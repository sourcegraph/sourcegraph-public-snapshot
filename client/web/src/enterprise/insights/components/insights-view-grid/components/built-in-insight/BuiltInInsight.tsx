import React, { forwardRef, useContext, useMemo, useState } from 'react'

import { useMergeRefs } from 'use-callback-ref'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, ParentSize, useDeepMemo } from '@sourcegraph/wildcard'

import { useSeriesToggle } from '../../../../../../insights/utils/use-series-toggle'
import { CodeInsightsBackendContext, LangStatsInsight } from '../../../../core'
import { InsightContentType } from '../../../../core/types/insight/common'
import { LazyQueryStatus } from '../../../../hooks/use-parallel-requests/use-parallel-request'
import { getTrackingTypeByInsightType, useCodeInsightViewPings } from '../../../../pings'
import {
    CategoricalBasedChartTypes,
    CategoricalChart,
    InsightCard,
    InsightCardBanner,
    InsightCardHeader,
    InsightCardLegend,
    InsightCardLoading,
    SeriesBasedChartTypes,
    SeriesChart,
} from '../../../views'
import { useInsightData } from '../../hooks/use-insight-data'
import { InsightContextMenu } from '../insight-context-menu/InsightContextMenu'
import { InsightContext } from '../InsightContext'

import styles from './BuiltInInsight.module.scss'

interface BuiltInInsightProps extends TelemetryProps, React.HTMLAttributes<HTMLElement> {
    insight: LangStatsInsight
    resizing: boolean
}

/**
 * Historically we had a few insights that were worked via extension API
 * search-based, code-stats insight
 *
 * This component renders insight card that works almost like before with extensions
 * Component sends FE network request to get and process information but does that in
 * main work thread instead of using Extension API.
 */
export const BuiltInInsight = forwardRef<HTMLElement, BuiltInInsightProps>((props, ref) => {
    const { insight, resizing, telemetryService, children, ...attributes } = props

    const { getBuiltInInsightData } = useContext(CodeInsightsBackendContext)
    const { currentDashboard, dashboards } = useContext(InsightContext)
    const seriesToggleState = useSeriesToggle()

    const cardRef = useMergeRefs([ref])
    const cachedInsight = useDeepMemo(insight)
    const { state, isVisible } = useInsightData(
        useMemo(() => () => getBuiltInInsightData({ insight: cachedInsight }), [getBuiltInInsightData, cachedInsight]),
        cardRef
    )

    // Visual line chart settings
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)

    const { trackDatumClicks, trackMouseLeave, trackMouseEnter } = useCodeInsightViewPings({
        telemetryService,
        insightType: getTrackingTypeByInsightType(insight.type),
    })

    return (
        <InsightCard
            {...attributes}
            ref={cardRef}
            data-testid={`insight-card.${insight.id}`}
            aria-label={`${insight.title} insight`}
            role="listitem"
            onMouseEnter={trackMouseEnter}
            onMouseLeave={trackMouseLeave}
        >
            <InsightCardHeader
                title={
                    <Link
                        to={`${window.location.origin}/insights/insight/${insight.id}`}
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        {insight.title}
                    </Link>
                }
            >
                {isVisible && (
                    <InsightContextMenu
                        insight={insight}
                        currentDashboard={currentDashboard}
                        dashboards={dashboards}
                        zeroYAxisMin={zeroYAxisMin}
                        onToggleZeroYAxisMin={() => setZeroYAxisMin(!zeroYAxisMin)}
                    />
                )}
            </InsightCardHeader>
            {resizing ? (
                <InsightCardBanner>Resizing</InsightCardBanner>
            ) : state.status === LazyQueryStatus.Loading || !isVisible ? (
                <InsightCardLoading>Loading code insight</InsightCardLoading>
            ) : state.status === LazyQueryStatus.Error ? (
                <ErrorAlert error={state.error} />
            ) : (
                <>
                    <ParentSize className={styles.chartContainer}>
                        {parent =>
                            state.data.type === InsightContentType.Series ? (
                                <SeriesChart
                                    type={SeriesBasedChartTypes.Line}
                                    width={parent.width}
                                    height={parent.height}
                                    zeroYAxisMin={zeroYAxisMin}
                                    locked={insight.isFrozen}
                                    className={styles.chart}
                                    onDatumClick={trackDatumClicks}
                                    seriesToggleState={seriesToggleState}
                                    {...state.data.content}
                                />
                            ) : (
                                <CategoricalChart
                                    type={CategoricalBasedChartTypes.Pie}
                                    width={parent.width}
                                    height={parent.height}
                                    locked={insight.isFrozen}
                                    className={styles.chart}
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
            {
                // Passing children props explicitly to render any top-level content like
                // resize-handler from the react-grid-layout library
                isVisible && children
            }
        </InsightCard>
    )
})
