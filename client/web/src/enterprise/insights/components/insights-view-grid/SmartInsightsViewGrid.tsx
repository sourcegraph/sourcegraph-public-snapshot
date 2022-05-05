import React, { memo, useCallback, useEffect, useState } from 'react'

import { isEqual } from 'lodash'
import { Layout, Layouts } from 'react-grid-layout'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ViewGrid } from '../../../../views'
import { Insight } from '../../core'
import { getTrackingTypeByInsightType } from '../../pings'

import { SmartInsight } from './components/SmartInsight'
import { insightLayoutGenerator, recalculateGridLayout } from './utils/grid-layout-generator'

interface SmartInsightsViewGridProps extends TelemetryProps {
    /**
     * List of built-in insights such as backend insight, FE search and code-stats
     * insights.
     */
    insights: Insight[]
    className?: string
}

/**
 * Renders grid of smart (stateful) insight card. These cards can independently extract and update
 * the insights settings (settings cascade subjects).
 */
export const SmartInsightsViewGrid: React.FunctionComponent<React.PropsWithChildren<SmartInsightsViewGridProps>> = memo(
    props => {
        const { telemetryService, insights, className } = props

        const [layouts, setLayouts] = useState<Layouts>({})
        const [resizingView, setResizeView] = useState<Layout | null>(null)

        useEffect(() => {
            setLayouts(insightLayoutGenerator(insights))
        }, [insights])

        const trackUICustomization = useCallback(
            (item: Layout) => {
                try {
                    const insight = insights.find(insight => item.i === insight.id)

                    if (insight) {
                        const insightType = getTrackingTypeByInsightType(insight.type)

                        telemetryService.log('InsightUICustomization', { insightType }, { insightType })
                    }
                } catch {
                    // noop
                }
            },
            [telemetryService, insights]
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
                setLayouts(recalculateGridLayout(allLayouts, insights))
            },
            [insights]
        )

        return (
            <ViewGrid
                layouts={layouts}
                className={className}
                onResizeStart={handleResizeStart}
                onResizeStop={handleResizeStop}
                onDragStart={trackUICustomization}
                onLayoutChange={handleLayoutChange}
            >
                {insights.map(insight => (
                    <SmartInsight
                        key={insight.id}
                        insight={insight}
                        telemetryService={telemetryService}
                        resizing={resizingView?.i === insight.id}
                    />
                ))}
            </ViewGrid>
        )
    },
    equalSmartGridProps
)

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
