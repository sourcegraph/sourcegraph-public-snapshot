import { ParentSize } from '@visx/responsive'
import * as H from 'history'
import React, { FunctionComponent, useCallback, useMemo } from 'react'
import { ChartContent } from 'sourcegraph'

import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { createProgrammaticLinkHandler } from '@sourcegraph/shared/src/util/link-click-handler/linkClickHandler'

import { BarChart } from './charts/bar/BarChart'
import { LineChart } from './charts/line/LineChart'
import { PieChart } from './charts/pie/PieChart'
import { DatumClickEvent } from './charts/types'

/**
 * Displays chart view content.
 */
export interface ChartViewContentProps {
    content: ChartContent
    history: H.History
    viewID: string
    telemetryService: TelemetryService
    className?: string
}

export const ChartViewContent: FunctionComponent<ChartViewContentProps> = props => {
    const { content, className = '', history, viewID, telemetryService } = props

    const handleDatumLinkClick = useCallback(
        () => {
            telemetryService.log('InsightDataPointClick', { insightType: viewID.split('.')[0] })
        },
        [viewID, telemetryService]
    )

    // Click link handler for line chart only. Catch click around point and redirect user by
    // link which we got from nearest datum to user cursor position
    const linkHandler = useMemo(() => {
        const linkHandler = createProgrammaticLinkHandler(history)
        return (event: DatumClickEvent): void => {

            console.log('click link datum', event);
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
                        return <LineChart {...content} onDatumClick={linkHandler} onDatumLinkClick={handleDatumLinkClick} width={width} height={height} />
                    }

                    if (content.chart === 'bar') {
                        return <BarChart {...content} width={width} height={height} onDatumClick={linkHandler} />
                    }

                    if (content.chart === 'pie') {
                        return <PieChart {...content} width={width} height={height} onDatumClick={linkHandler} />
                    }

                    // TODO Add UI for incorrect type of chart
                    return null
                }}
            </ParentSize>
        </div>
    )
}
