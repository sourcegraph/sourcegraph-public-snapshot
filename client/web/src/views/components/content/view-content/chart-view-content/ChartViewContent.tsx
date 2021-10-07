import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import React, { FunctionComponent, useCallback } from 'react'
import { ChartContent } from 'sourcegraph'

import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { BarChart } from './charts/bar/BarChart'
import { LineChart } from './charts/line/LineChart'
import { DatumZoneClickEvent } from './charts/line/types'
import { PieChart } from './charts/pie/PieChart'
import { getInsightTypeByViewId } from './utils/get-insight-type-by-view-id'

export interface ChartViewContentProps {
    /** Data for chart (lines, bar, pie arcs)*/
    content: ChartContent
    /** Extension view ID*/
    viewID: string
    telemetryService: TelemetryService
    className?: string
}

/**
 * Display chart content with different type of charts (line, bar, pie)
 */
export const ChartViewContent: FunctionComponent<ChartViewContentProps> = props => {
    const { content, className = '', viewID, telemetryService } = props

    const handleDatumLinkClick = useCallback((): void => {
        telemetryService.log(
            'InsightDataPointClick',
            { insightType: getInsightTypeByViewId(viewID) },
            { insightType: getInsightTypeByViewId(viewID) }
        )
    }, [viewID, telemetryService])

    // Click link-zone handler for line chart only. Catch click around point and redirect user by
    // link which we've got from nearest datum point to user cursor position. This allows user
    // not to aim on small point on the chart and just click somewhere around the point.
    const linkHandler = useCallback(
        (event: DatumZoneClickEvent): void => {
            if (!event.link) {
                return
            }

            telemetryService.log(
                'InsightDataPointClick',
                { insightType: getInsightTypeByViewId(viewID) },
                { insightType: getInsightTypeByViewId(viewID) }
            )

            window.open(event.link, '_blank', 'noopener')?.focus()
        },
        [viewID, telemetryService]
    )

    return (
        <div className={classNames('chart-view-content', className)}>
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
