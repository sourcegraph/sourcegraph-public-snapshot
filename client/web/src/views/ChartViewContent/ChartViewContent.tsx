import React, { FunctionComponent, useMemo } from 'react'
import { ChartContent } from 'sourcegraph'
import * as H from 'history'
import { ParentSize } from '@visx/responsive'
import { createProgrammaticallyLinkHandler } from '@sourcegraph/shared/src/util/link-click-handler/linkClickHandler'

import { eventLogger } from '../../tracking/eventLogger'
import { LineChart } from './charts/line/LineChart'
import { PieChart } from './charts/pie/PieChart'
import { BarChart } from './charts/bar/BarChart'
import { DatumClickEvent } from './charts/types'

/**
 * Displays chart view content.
 */
export interface ChartViewContentProps {
    content: ChartContent
    history: H.History
    viewID: string
    className?: string
}

export const ChartViewContent: FunctionComponent<ChartViewContentProps> = props => {
    const { content, className = '', history, viewID } = props

    const linkHandler = useMemo(() => {
        const linkHandler = createProgrammaticallyLinkHandler(history)
        return (event: DatumClickEvent): void => {
            if (!event.link) {
                return
            }

            eventLogger.log('InsightDataPointClick', { insightType: viewID.split('.')[0] })
            linkHandler(event.originEvent, event.link)
        }
    }, [history, viewID])

    return (
        <div className={`chart-view-content ${className}`}>
            <ParentSize className="chart-view-content__chart">
                {({ width, height }) => {
                    if (content.chart === 'bar') {
                        return <BarChart {...content} width={width} height={height} onDatumClick={linkHandler} />
                    }

                    if (content.chart === 'line') {
                        return <LineChart {...content} width={width} height={height} onDatumClick={linkHandler} />
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
