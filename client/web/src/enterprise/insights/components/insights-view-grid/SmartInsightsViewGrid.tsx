import { isEqual } from 'lodash'
import React, { memo, useCallback, useEffect, useMemo, useState } from 'react'
import { Layout, Layouts } from 'react-grid-layout'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ViewGrid } from '../../../../views'
import { Insight, InsightsDashboardScope } from '../../core/types'
import { getTrackingTypeByInsightType } from '../../pings'

import { LockedBanner } from './components/locked-banner/LockedBanner'
import { SmartInsight } from './components/smart-insight/SmartInsight'
import { insightLayoutGenerator, recalculateGridLayout } from './utils/grid-layout-generator'

interface SmartInsightsViewGridProps extends TelemetryProps {
    /**
     * List of built-in insights such as backend insight, FE search and code-stats
     * insights.
     */
    insights: Insight[]
    dashboardScope: InsightsDashboardScope
}

const INSIGHT_PAGE_CONTEXT = {}

/**
 * Renders grid of smart (stateful) insight card. These cards can independently extract and update
 * the insights settings (settings cascade subjects).
 */
export const SmartInsightsViewGrid: React.FunctionComponent<SmartInsightsViewGridProps> = memo(props => {
    const { telemetryService, insights, dashboardScope } = props

    const [layouts, setLayouts] = useState<Layouts>({})
    const [resizingView, setResizeView] = useState<Layout | null>(null)

    // TODO: remove this once backend starts providing the "locked" value
    const parsedInsights = useMemo(
        () =>
            insights.map((insight, index) => ({
                ...insight,
                locked: dashboardScope !== InsightsDashboardScope.Global || index > 2,
            })),
        [insights, dashboardScope]
    )

    useEffect(() => {
        setLayouts(insightLayoutGenerator(parsedInsights))
    }, [parsedInsights])

    const trackUICustomization = useCallback(
        (item: Layout) => {
            try {
                const insight = parsedInsights.find(insight => item.i === insight.id)

                if (insight) {
                    const insightType = getTrackingTypeByInsightType(insight.viewType)

                    telemetryService.log('InsightUICustomization', { insightType }, { insightType })
                }
            } catch {
                // noop
            }
        },
        [telemetryService, parsedInsights]
    )

    const handleResizeStart = useCallback(
        (item: Layout) => {
            setResizeView(item)
            trackUICustomization(item)
        },
        [trackUICustomization]
    )

    const handleResizeStop = useCallback((item: Layout) => {
        setResizeView(null)
    }, [])

    const handleLayoutChange = useCallback(
        (currentLayout: Layout[], allLayouts: Layouts): void => {
            setLayouts(recalculateGridLayout(allLayouts, parsedInsights))
        },
        [parsedInsights]
    )

    return (
        <ViewGrid
            layouts={layouts}
            onResizeStart={handleResizeStart}
            onResizeStop={handleResizeStop}
            onDragStart={trackUICustomization}
            onLayoutChange={handleLayoutChange}
        >
            {parsedInsights.map(insight => (
                <SmartInsight
                    key={insight.id}
                    insight={insight}
                    telemetryService={telemetryService}
                    resizing={resizingView?.i === insight.id}
                    // Set execution insight context explicitly since this grid component is used
                    // only for the dashboard (insights) page
                    where="insightsPage"
                    context={INSIGHT_PAGE_CONTEXT}
                    alternate={insight.locked ? <LockedBanner /> : undefined}
                />
            ))}
        </ViewGrid>
    )
}, equalSmartGridProps)

/**
 * Custom props checker for the smart grid component.
 *
 * Ignore settings cascade change and insight body config changes to avoid
 * animations of grid item rerender and grid position items. In some cases (like insight
 * filters updating, we want to ignore insights from settings cascade).
 * But still trigger grid animation rerender if insight ordering or insight count
 * have been changed.
 */
function equalSmartGridProps(
    previousProps: SmartInsightsViewGridProps,
    nextProps: SmartInsightsViewGridProps
): boolean {
    const { insights: previousInsights, ...otherPrepProps } = previousProps
    const { insights: nextInsights, ...otherNextProps } = nextProps

    if (!isEqual(otherPrepProps, otherNextProps)) {
        return false
    }

    return isEqual(
        previousInsights.map(insight => insight.id),
        nextInsights.map(insight => insight.id)
    )
}
