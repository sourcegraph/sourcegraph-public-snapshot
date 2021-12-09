import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import React, { FunctionComponent, useCallback } from 'react'
import { ChartContent } from 'sourcegraph'

import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { BarChart } from './charts/bar/BarChart'
import { LineChart } from './charts/line'
import { DatumZoneClickEvent } from './charts/line/types'
import { PieChart } from './charts/pie/PieChart'
import styles from './ChartViewContent.module.scss'
import { getInsightTypeByViewId } from './utils/get-insight-type-by-view-id'

export enum ChartViewContentLayout {
    /**
     * With this layout chart takes all available space inside parent
     * block and tries to fit chart content to this rectangle.
     *
     * This by default mode is used in code insights grid layout.
     */
    ByParentSize,

    /**
     * With this layout chart render content and if expand parent
     * block if it's too small to fit all content.
     */
    ByContentSize,
}

export interface ChartViewContentProps {
    /** Data for chart (lines, bar, pie arcs)*/
    content: ChartContent
    /** Extension view ID*/
    viewID: string
    telemetryService: TelemetryService
    className?: string
    layout?: ChartViewContentLayout
}

/**
 * Display chart content with different type of charts (line, bar, pie)
 */
export const ChartViewContent: FunctionComponent<ChartViewContentProps> = props => {
    const { content, className = '', viewID, telemetryService, layout = ChartViewContentLayout.ByParentSize } = props

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

    const isResponsive = layout === ChartViewContentLayout.ByParentSize

    return (
        <div className={classNames(styles.chartViewContent, className)}>
            <ParentSize
                data-chart-size-root=""
                className={classNames(styles.chart, { [styles.chartWithResponsive]: isResponsive })}
            >
                {({ width, height }) => {
                    if (content.chart === 'line') {
                        return (
                            <LineChart
                                {...content}
                                onDatumZoneClick={linkHandler}
                                onDatumLinkClick={handleDatumLinkClick}
                                width={width}
                                height={height}
                                hasChartParentFixedSize={isResponsive}
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
