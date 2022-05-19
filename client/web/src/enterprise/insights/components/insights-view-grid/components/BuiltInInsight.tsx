import React, { Ref, useContext, useMemo, useRef, useState } from 'react'

import { useMergeRefs } from 'use-callback-ref'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useDeepMemo } from '@sourcegraph/wildcard'

import { ParentSize } from '../../../../../charts'
import { CodeInsightsBackendContext, LangStatsInsight, SearchRuntimeBasedInsight } from '../../../core'
import { InsightContentType } from '../../../core/types/insight/common'
import { LazyQueryStatus } from '../../../hooks/use-parallel-requests/use-parallel-request'
import { getTrackingTypeByInsightType, useCodeInsightViewPings } from '../../../pings'
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
} from '../../views'
import { useInsightData } from '../hooks/use-insight-data'

import { InsightContextMenu } from './insight-context-menu/InsightContextMenu'
import { InsightContext } from './InsightContext'

interface BuiltInInsightProps extends TelemetryProps, React.HTMLAttributes<HTMLElement> {
    insight: SearchRuntimeBasedInsight | LangStatsInsight
    innerRef: Ref<HTMLElement>
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
export function BuiltInInsight(props: BuiltInInsightProps): React.ReactElement {
    const { insight, resizing, telemetryService, innerRef, ...otherProps } = props
    const { getBuiltInInsightData } = useContext(CodeInsightsBackendContext)
    const { currentDashboard, dashboards } = useContext(InsightContext)

    const insightCardReference = useRef<HTMLDivElement>(null)
    const mergedInsightCardReference = useMergeRefs([insightCardReference, innerRef])

    const cachedInsight = useDeepMemo(insight)

    const { state, isVisible } = useInsightData(
        useMemo(() => () => getBuiltInInsightData({ insight: cachedInsight }), [getBuiltInInsightData, cachedInsight]),
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
            {...otherProps}
            ref={mergedInsightCardReference}
            data-testid={`insight-card.${insight.id}`}
            onMouseEnter={trackMouseEnter}
            onMouseLeave={trackMouseLeave}
        >
            <InsightCardHeader title={insight.title}>
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
                    <ParentSize>
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
            {
                // Passing children props explicitly to render any top-level content like
                // resize-handler from the react-grid-layout library
                isVisible && otherProps.children
            }
        </InsightCard>
    )
}
