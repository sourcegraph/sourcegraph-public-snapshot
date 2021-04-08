import { ParentSize } from '@visx/responsive'
import * as H from 'history'
import React, { FunctionComponent, useCallback, useMemo } from 'react'
import { ChartContent } from 'sourcegraph'

import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { createProgrammaticLinkHandler } from '@sourcegraph/shared/src/util/link-click-handler/linkClickHandler'

import { BarChart } from './charts/bar/BarChart'
import { LineChart } from './charts/line/LineChart'
import { PieChart } from './charts/pie/PieChart'
import { DatumZoneClickEvent } from './charts/line/types'

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

    const handleDatumLinkClick = useCallback(() => {
        telemetryService.log('InsightDataPointClick', { insightType: viewID.split('.')[0] })
    }, [viewID, telemetryService])

    // Click link-zone handler for line chart only. Catch click around point and redirect user by
    // link which we've got from nearest datum point to user cursor position.
    // This allows user not to aim on small point on the chart and just click somewhere around point.
    const linkHandler = useMemo(() => {
        const linkHandler = createProgrammaticLinkHandler(history)
        return (event: DatumZoneClickEvent): void => {
            if (!event.link) {
                return
            }

            telemetryService.log('InsightDataPointClick', { insightType: viewID.split('.')[0] })
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
