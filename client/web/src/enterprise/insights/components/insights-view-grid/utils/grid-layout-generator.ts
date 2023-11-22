import type { Layout, Layouts as ReactGridLayouts } from 'react-grid-layout'

import {
    type CaptureGroupInsight,
    type Insight,
    isCaptureGroupInsight,
    isComputeInsight,
    isSearchBasedInsight,
    type SearchBasedInsight,
} from '../../../core'
import {
    type BreakpointName,
    BREAKPOINTS_NAMES,
    COLUMNS,
    DEFAULT_HEIGHT,
    DEFAULT_ITEMS_PER_ROW,
    MIN_WIDTHS,
} from '../components/view-grid/ViewGrid'

const MINIMAL_SERIES_FOR_ASIDE_LEGEND = 4
const MIN_WIDTHS_LANDSCAPE_MODE: Record<BreakpointName, number> = { xs: 1, sm: 3, md: 4, lg: 4 }

type InsightWithLegend = SearchBasedInsight | CaptureGroupInsight

const isManySeriesInsight = (insight: Insight): insight is InsightWithLegend =>
    isCaptureGroupInsight(insight) ||
    isComputeInsight(insight) ||
    (isSearchBasedInsight(insight) && insight.series.length >= MINIMAL_SERIES_FOR_ASIDE_LEGEND)

const getMinWidth = (breakpoint: BreakpointName, insight: Insight): number =>
    isManySeriesInsight(insight) ? MIN_WIDTHS_LANDSCAPE_MODE[breakpoint] : MIN_WIDTHS[breakpoint]

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
export const insightLayoutGenerator = (
    insights: Insight[],
    persistedLayouts: ReactGridLayouts | null
): ReactGridLayouts => {
    return Object.fromEntries(
        BREAKPOINTS_NAMES.map(
            breakpointName => [breakpointName, generateLayout(breakpointName, persistedLayouts)] as const
        )
    )

    function generateLayout(breakpointName: BreakpointName, persistedLayouts: ReactGridLayouts | null): Layout[] {
        switch (breakpointName) {
            case 'lg': {
                return generateComplexLayout(insights, breakpointName, persistedLayouts?.lg ?? [])
            }
            default: {
                return generatePlainLayout(insights, breakpointName)
            }
        }
    }
}

function generatePlainLayout(insights: Insight[], breakpointName: BreakpointName): Layout[] {
    return insights.map((insight, index) => {
        const width = COLUMNS[breakpointName] / DEFAULT_ITEMS_PER_ROW[breakpointName]

        return {
            i: insight.id,
            // Increase height of chart block if view has many data series
            h: DEFAULT_HEIGHT,
            w: width,
            x: (index * width) % COLUMNS[breakpointName],
            y: Math.floor((index * width) / COLUMNS[breakpointName]),
            minW: getMinWidth(breakpointName, insight),
            minH: DEFAULT_HEIGHT,
        }
    })
}

function generateComplexLayout(
    insights: Insight[],
    breakpointName: BreakpointName,
    persistedLayouts: Layout[]
): Layout[] {
    // Filter out insights that we don't have in persisted layouts grid already
    const newInsights = insights.filter(insight => !persistedLayouts.find(lsInsight => lsInsight.i === insight.id))

    // Filter out layouts that we don't have in the most recent list of insights,
    // like in case if some insight has been deleted from dashboard we shouldn't
    // generate a card item for it anymore
    const existingLayouts = persistedLayouts.filter(insight => insights.find(item => item.id === insight.i))

    // Calculate the Y coordinate for new row (there will be added all new insights
    // that we don't have in the persisted insight layouts)
    const persistedInsightsY = existingLayouts.reduce((maxY, layout) => Math.max(maxY, layout.y), 0) + 1

    const newLayouts = newInsights
        .reduce<Layout[][]>(
            (grid, insight) => {
                const itemsPerRow = isManySeriesInsight(insight) ? 2 : DEFAULT_ITEMS_PER_ROW[breakpointName]
                const columnsPerRow = COLUMNS[breakpointName]
                const width = columnsPerRow / itemsPerRow
                const lastRow = grid.at(-1)!
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
                            y: persistedInsightsY + grid.length,
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
                        y: persistedInsightsY + grid.length - 1,
                        minW: getMinWidth(breakpointName, insight),
                        minH: DEFAULT_HEIGHT,
                    })
                }

                return grid
            },
            [[]]
        )
        .flat()

    return [...existingLayouts, ...newLayouts]
}
