import { ParentSize } from '@visx/responsive'
import * as H from 'history'
import React, { FunctionComponent, useMemo } from 'react'
import { ChartContent } from 'sourcegraph'

import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    createLinkClickHandler,
    createProgrammaticLinkHandler,
} from '@sourcegraph/shared/src/util/link-click-handler/linkClickHandler'

import { BarChart } from './charts/bar/BarChart'
import { LineChart } from './charts/line/LineChart'
import { DatumZoneClickEvent } from './charts/line/types'
import { PieChart } from './charts/pie/PieChart'
import { getInsightTypeByViewId } from './utils/get-insight-type-by-view-id'

/**
 * Displays chart view content.
 */
export interface ChartViewContentProps {
    /** Data for chart (lines, bar, pie arcs)*/
    content: ChartContent
    /** History object to redirect user if he clicked on bar point or arc with link */
    history: H.History
    /** Extension view ID*/
    viewID: string
    telemetryService: TelemetryService
    className?: string
}

/**
 * Display chart content with different type of charts (line, bar, pie)
 */
export const ChartViewContent: FunctionComponent<ChartViewContentProps> = props => {
    const { content, className = '', history, viewID, telemetryService } = props

    const handleDatumLinkClick = useMemo(() => {
        const nativeLinkHandler = createLinkClickHandler(history)

        return (event: React.MouseEvent) => {
            telemetryService.log('InsightDataPointClick', { insightType: getInsightTypeByViewId(viewID) })
            nativeLinkHandler(event)
        }
    }, [viewID, telemetryService, history])

    // Click link-zone handler for line chart only. Catch click around point and redirect user by
    // link which we've got from nearest datum point to user cursor position. This allows user
    // not to aim on small point on the chart and just click somewhere around the point.
    const linkHandler = useMemo(() => {
        const linkHandler = createProgrammaticLinkHandler(history)
        return (event: DatumZoneClickEvent): void => {
            if (!event.link) {
                return
            }

            telemetryService.log('InsightDataPointClick', { insightType: getInsightTypeByViewId(viewID) })
            linkHandler(event.originEvent, event.link)
        }
    }, [history, viewID, telemetryService])

    return (
        <div className={`chart-view-content ${className}`}>
            <ParentSize className="chart-view-content__chart">
                {({ width, height }) => {
                    if (content.chart === 'line') {
                        return (
                            <LineChart
                                {...content}
                                onDatumZoneClick={linkHandler}
                                onDatumLinkClick={handleDatumLinkClick}
                                width={width}
                                height={height}
                            />
                        )
                    }

                    if (content.chart === 'bar') {
                        return (
                            <BarChart
                                {...content}
                                width={width}
                                height={height}
                                onDatumLinkClick={handleDatumLinkClick}
                            />
                        )
                    }

                    if (content.chart === 'pie') {
                        return (
                            <PieChart
                                {...content}
                                width={width}
                                height={height}
                                onDatumLinkClick={handleDatumLinkClick}
                            />
                        )
                    }

                    // TODO Add UI for incorrect type of chart
                    return null
                }}
            </ParentSize>
        </div>
    )
}
