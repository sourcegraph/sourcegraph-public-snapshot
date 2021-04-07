import React, { ReactElement } from 'react'
import { ParentSize } from '@visx/responsive'

import { DEFAULT_LINE_STROKE } from './colors'
import { LineChartContent, LineChartContentProps } from './components/LineChartContent'

export interface LineChartProps<Datum extends object> extends LineChartContentProps<Datum> {}

/**
 * Display responsive line chart with legend below the chart.
 */
export function LineChart<Datum extends object>(props: LineChartProps<Datum>): ReactElement {
    const { width, height, ...otherProps } = props
    const hasLegend = props.series.every(line => !!line.name)

    if (!hasLegend) {
        return <LineChartContent {...props} />
    }

    return (
        /* eslint-disable-next-line react/forbid-dom-props */
        <div style={{ width, height }} className="line-chart">
            {/*
                In case if we have a legend to render we have to have responsive container for chart
                just to calculate right sizes for chart content = rootContainerSizes - legendSizes
            */}
            <ParentSize className="line-chart__content-parent-size">
                {({ width, height }) => <LineChartContent {...otherProps} width={width} height={height} />}
            </ParentSize>

            <ul className="line-chart__legend">
                {props.series.map(line => (
                    <li key={line.dataKey.toString()} className="line-chart__legend-item">
                        <div
                            /* eslint-disable-next-line react/forbid-dom-props */
                            style={{ backgroundColor: line.stroke ?? DEFAULT_LINE_STROKE }}
                            className="line-chart__legend-mark"
                        />
                        {line.name}
                    </li>
                ))}
            </ul>
        </div>
    )
}
