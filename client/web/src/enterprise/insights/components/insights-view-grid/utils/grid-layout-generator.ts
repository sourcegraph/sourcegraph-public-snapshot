import { Layout, Layouts as ReactGridLayouts } from 'react-grid-layout'

import {
    CaptureGroupInsight,
    Insight,
    isCaptureGroupInsight,
    isComputeInsight,
    isSearchBasedInsight,
    SearchBasedInsight,
} from '../../../core'
import {
    BreakpointName,
    BREAKPOINTS_NAMES,
    COLUMNS,
    DEFAULT_HEIGHT,
    DEFAULT_ITEMS_PER_ROW,
    MIN_WIDTHS,
} from '../components/view-grid/ViewGrid'

const MINIMAL_SERIES_FOR_ASIDE_LEGEND = 3
const MIN_WIDTHS_LANDSCAPE_MODE: Record<BreakpointName, number> = { xs: 1, sm: 3, md: 4, lg: 4 }

type InsightWithLegend = SearchBasedInsight | CaptureGroupInsight

const isManySeriesInsight = (insight: Insight): insight is InsightWithLegend =>
    isCaptureGroupInsight(insight) ||
    isComputeInsight(insight) ||
    (isSearchBasedInsight(insight) && insight.series.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND)

const getMinWidth = (breakpoint: BreakpointName, insight: Insight): number =>
    isManySeriesInsight(insight) ? MIN_WIDTHS_LANDSCAPE_MODE[breakpoint] : MIN_WIDTHS[breakpoint]

const getMinHeight = (insight: Insight): number => {
    if (!isManySeriesInsight(insight)) {
        return DEFAULT_HEIGHT
    }

    return isSearchBasedInsight(insight)
        ? Math.min(DEFAULT_HEIGHT + insight.series.length * 0.1, DEFAULT_HEIGHT * 2)
        : DEFAULT_HEIGHT * 2
}

/**
 * Custom Code Insight Grid layout generator. For different screens (xs, sm, md, lg) it
 * generates different initial layouts. See examples below
 *
 * <pre>
 * Large break points (lg)                         Mid size and small breakpoints (xs, sm, md)
 * ┌────────────┐ ┌────────────┐ ┌────────────┐    ┌────────────┐ ┌────────────┐ ┌────────────┐
 * │▪▪▪▪▪▪▪▪▪   │ │▪▪▪▪▪       │ │▪▪▪▪▪▪▪▪▪   │    │▪▪▪▪▪▪▪▪▪   │ │▪▪▪▪▪       │ │▪▪▪▪▪▪▪▪▪   │
 * │            │ │            │ │            │    │            │ │            │ │            │
 * │            │ │            │ │            │    │            │ │            │ │            │
 * │            │ │            │ │            │    │            │ │            │ │            │
 * │           ◿│ │           ◿│ │           ◿│    │           ◿│ │           ◿│ │           ◿│
 * └────────────┘ └────────────┘ └────────────┘    └────────────┘ └────────────┘ └────────────┘
 * ┌────────────────────┐┌────────────────────┐    ┌────────────┐ ┌────────────┐ ┌────────────┐
 * │■■■■■■■■■■■■■       ││■■■■■■■■■           │    │■■■■■■■■■■■■│ │▪▪▪▪▪▪▪▪▪   │ │▪▪▪▪▪       │
 * │                    ││                    │    │            │ │            │ │            │
 * │ Insight with 3 and ││Insight with 3 and  │    │ Insight    │ │            │ │            │
 * │ more series        ││more series         │    │ with 3 and │ │            │ │            │
 * │                   ◿││                   ◿│    │ more       │ │           ◿│ │           ◿│
 * └────────────────────┘└────────────────────┘    │            │ └────────────┘ └────────────┘
 * ┌────────────┐ ┌────────────┐                   │            │ ┌────────────┐ ┌────────────┐
 * │▪▪▪▪▪▪▪▪▪   │ │▪▪▪▪▪       │                   │            │ │▪▪▪▪▪▪▪▪▪   │ │▪▪▪▪▪▪▪▪▪   │
 * │            │ │            │                   └────────────┘ │            │ │            │
 * │            │ │            │                                  │            │ │            │
 * │            │ │            │                                  │            │ │            │
 * │           ◿│ │           ◿│                                  │           ◿│ │           ◿│
 * └────────────┘ └────────────┘                                  └────────────┘ └────────────┘
 * </pre>
 */
export const insightLayoutGenerator = (insights: Insight[]): ReactGridLayouts => {
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
                        // Increase height of chart block if view has many data series
                        h: getMinHeight(insight),
                        w: width,
                        x: (index * width) % COLUMNS[breakpointName],
                        y: Math.floor((index * width) / COLUMNS[breakpointName]),
                        minW: getMinWidth(breakpointName, insight),
                        minH: getMinHeight(insight),
                    }
                })
            }

            case 'lg': {
                return insights
                    .reduce<Layout[][]>(
                        (grid, insight) => {
                            const itemsPerRow = isManySeriesInsight(insight) ? 2 : DEFAULT_ITEMS_PER_ROW[breakpointName]
                            const columnsPerRow = COLUMNS[breakpointName]
                            const width = columnsPerRow / itemsPerRow
                            const lastRow = grid[grid.length - 1]
                            const lastRowCurrentWidth = lastRow.reduce((sumWidth, element) => sumWidth + element.w, 0)

                            // Move element on new line (row)
                            if (lastRowCurrentWidth + width > columnsPerRow) {
                                // Adjust elements width on the same row if no more elements don't
                                // fit in this row
                                for (const [index, element] of lastRow.entries()) {
                                    element.w = columnsPerRow / lastRow.length
                                    element.x = (index * columnsPerRow) / lastRow.length
                                }

                                // Create new row
                                grid.push([
                                    {
                                        i: insight.id,
                                        h: DEFAULT_HEIGHT,
                                        w: width,
                                        x: 0,
                                        y: grid.length,
                                        minW: getMinWidth(breakpointName, insight),
                                        minH: DEFAULT_HEIGHT,
                                    },
                                ])
                            } else {
                                // Add another element to the last row of the grid
                                lastRow.push({
                                    i: insight.id,
                                    h: DEFAULT_HEIGHT,
                                    w: width,
                                    x: lastRowCurrentWidth,
                                    y: grid.length - 1,
                                    minW: getMinWidth(breakpointName, insight),
                                    minH: DEFAULT_HEIGHT,
                                })
                            }

                            return grid
                        },
                        [[]]
                    )
                    .flat()
            }
        }
    }
}

export const recalculateGridLayout = (nextLayouts: ReactGridLayouts, insights: Insight[]): ReactGridLayouts => {
    const keys = Object.keys(nextLayouts) as unknown as BreakpointName[]
    const insightsMap = Object.fromEntries(insights.map(insight => [insight.id, insight]))
    const adjustedLayouts: ReactGridLayouts = {}

    for (const key of keys) {
        const layout = nextLayouts[key]

        adjustedLayouts[key] = layout.map(item => {
            const insight = insightsMap[item.i]

            if (!insight) {
                return item
            }

            if (isManySeriesInsight(insight) && item.minW === item.w) {
                item.minH = getMinHeight(insight)
                item.h = item.h > (item.minH ?? 0) ? item.h : getMinHeight(insight)
            } else {
                item.minH = DEFAULT_HEIGHT
            }

            item.h = item.minH > item.h ? item.minH : item.h

            return item
        })
    }

    return adjustedLayouts
}
