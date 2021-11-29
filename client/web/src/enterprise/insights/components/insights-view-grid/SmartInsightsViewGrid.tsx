import { isEqual } from 'lodash'
import React, { memo } from 'react'
import { Layout, Layouts as ReactGridLayouts } from 'react-grid-layout'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import {
    BreakpointName,
    BREAKPOINTS_NAMES,
    COLUMNS,
    DEFAULT_HEIGHT,
    DEFAULT_ITEMS_PER_ROW,
    MIN_WIDTHS,
    ViewGrid,
} from '../../../../views'
import { Insight, isSearchBasedInsight } from '../../core/types'

import { SmartInsight } from './components/smart-insight/SmartInsight'

const insightLayoutGenerator = (insights: Insight[]): ReactGridLayouts => {

    return Object.fromEntries(
        BREAKPOINTS_NAMES.map(breakpointName => [breakpointName, generateLayout(breakpointName)] as const)
    )

    function generateLayout(breakpointName: BreakpointName): Layout[] {
        switch (breakpointName) {
            case 'xs':
            case 'sm':
            case 'md': {
                return insights.map((insight, index) => {
                    const width = COLUMNS[breakpointName] / DEFAULT_ITEMS_PER_ROW[breakpointName]
                    return {
                        i: insight.id,
                        h: DEFAULT_HEIGHT,
                        w: width,
                        x: (index * width) % COLUMNS[breakpointName],
                        y: Math.floor((index * width) / COLUMNS[breakpointName]),
                        minW: MIN_WIDTHS[breakpointName],
                        minH: 2,
                    }
                })
            }

            case 'lg': {
                return insights.reduce<Layout[][]>((grid, insight) => {
                    const isManySeriesChart = isSearchBasedInsight(insight) && insight.series.length > 3
                    const itemsPerRow = isManySeriesChart ? 2 : DEFAULT_ITEMS_PER_ROW[breakpointName]
                    const columnsPerRow = COLUMNS[breakpointName]
                    const width = columnsPerRow / itemsPerRow
                    const lastRow = grid[grid.length - 1]
                    const lastRowCurrentWidth = lastRow.reduce((sumWidth, element) => sumWidth + element.w, 0)

                    // Move element on new line (row)
                    if (lastRowCurrentWidth + width > columnsPerRow) {
                        const newRow = [
                            {
                                i: insight.id,
                                h: DEFAULT_HEIGHT,
                                w: width,
                                x: 0,
                                y: grid.length,
                                minW: MIN_WIDTHS[breakpointName],
                                minH: 2,
                            }
                        ]

                        for (const [index, element] of lastRow.entries()) {
                            element.w = columnsPerRow / lastRow.length
                            element.x = index * columnsPerRow / lastRow.length
                        }

                        grid.push(newRow)
                    } else {
                        lastRow.push({
                            i: insight.id,
                            h: DEFAULT_HEIGHT,
                            w: width,
                            x: lastRowCurrentWidth,
                            y: grid.length - 1,
                            minW: MIN_WIDTHS[breakpointName],
                            minH: 2,
                        })
                    }

                    return grid
                }, [[]]).flat()
            }
        }
    }
}

interface SmartInsightsViewGridProps extends TelemetryProps {
    /**
     * List of built-in insights such as backend insight, FE search and code-stats
     * insights.
     */
    insights: Insight[]
}

const INSIGHT_PAGE_CONTEXT = {}

/**
 * Renders grid of smart (stateful) insight card. These cards can independently extract and update
 * the insights settings (settings cascade subjects).
 */
export const SmartInsightsViewGrid: React.FunctionComponent<SmartInsightsViewGridProps> = memo(props => {
    const { telemetryService, insights } = props

    return (
        <ViewGrid
            layouts={insightLayoutGenerator(insights)}
            telemetryService={telemetryService}
         >
            {insights.map(insight => (
                <SmartInsight
                    key={insight.id}
                    insight={insight}
                    telemetryService={telemetryService}
                    // Set execution insight context explicitly since this grid component is used
                    // only for the dashboard (insights) page
                    where="insightsPage"
                    context={INSIGHT_PAGE_CONTEXT}
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
