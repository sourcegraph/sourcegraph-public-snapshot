import { ParentSize } from '@visx/responsive'
import { EventEmitterProvider } from '@visx/xychart'
import classnames from 'classnames'
import React, { ReactElement } from 'react'

import { getLineStroke, LineChartContent, LineChartContentProps } from './components/LineChartContent'

/**
 * Check percy test run to be able disable flaky line chart tooltip appearance
 * by disabling any point events over line chart container.
 * See https://github.com/sourcegraph/sourcegraph/issues/23669
 */
const IS_PERCY_RUN = process.env.PERCY_ON === 'true'

export interface LineChartProps<Datum extends object> extends LineChartContentProps<Datum> {}

/**
 * Display responsive line chart with legend below the chart.
 */
export function LineChart<Datum extends object>(props: LineChartProps<Datum>): ReactElement {
    const { width, height, ...otherProps } = props
    const hasLegend = props.series.every(line => !!line.name)

    if (!hasLegend) {
        return (
            // Because we need to catch all events from line chart by ourselves we have to
            // use this chart's event emitter for override some events handler and bind them
            // to custom elements within LineChartContent component.
            <EventEmitterProvider>
                <LineChartContent {...props} />
            </EventEmitterProvider>
        )
    }

    return (
        <EventEmitterProvider>
            <div
                aria-label="Line chart"
                /* eslint-disable-next-line react/forbid-dom-props */
                style={{ width, height }}
                className={classnames('line-chart', { 'line-chart--no-interaction': IS_PERCY_RUN })}
            >
                {/*
                    In case if we have a legend to render we have to have responsive container for chart
                    just to calculate right sizes for chart content = rootContainerSizes - legendSizes
                */}
                <ParentSize className="line-chart__content-parent-size">
                    {({ width, height }) => <LineChartContent {...otherProps} width={width} height={height} />}
                </ParentSize>

                <ul aria-hidden={true} className="line-chart__legend">
                    {props.series.map(line => (
                        <li key={line.dataKey.toString()} className="line-chart__legend-item">
                            <div
                                /* eslint-disable-next-line react/forbid-dom-props */
                                style={{ backgroundColor: getLineStroke(line) }}
                                className="line-chart__legend-mark"
                            />
                            {line.name}
                        </li>
                    ))}
                </ul>
            </div>
        </EventEmitterProvider>
    )
}
