import { ParentSize } from '@visx/responsive'
import { EventEmitterProvider } from '@visx/xychart'
import classNames from 'classnames'
import React, { ReactElement, useContext } from 'react'

import { getLineStroke, LineChartContent, LineChartContentProps } from './components/LineChartContent'
import { LineChartSettingsContext } from './line-chart-settings-provider'
import styles from './LineChart.module.scss'

export interface LineChartProps<Datum extends object> extends LineChartContentProps<Datum> {}

/**
 * Display responsive line chart with legend below the chart.
 */
export function LineChart<Datum extends object>(props: LineChartProps<Datum>): ReactElement {
    const { width, height, ...otherProps } = props
    const { layout } = useContext(LineChartSettingsContext)
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

    const isHorizontal = layout === 'horizontal'

    return (
        <EventEmitterProvider>
            <div
                aria-label="Line chart"
                /* eslint-disable-next-line react/forbid-dom-props */
                style={{ width, height }}
                className={classNames(styles.lineChart, { [styles.lineChartHorizontal]: isHorizontal })}
            >
                {/*
                    In case if we have a legend to render we have to have responsive container for chart
                    just to calculate right sizes for chart content = rootContainerSizes - legendSizes
                */}
                <ParentSize className={styles.contentParentSize}>
                    {({ width, height }) => <LineChartContent {...otherProps} width={width} height={height} />}
                </ParentSize>

                <ul
                    aria-hidden={true}
                    className={classNames(styles.legend, { [styles.legendHorizontal]: isHorizontal })}
                >
                    {props.series.map(line => (
                        <li key={line.dataKey.toString()} className={styles.legendItem}>
                            <div
                                /* eslint-disable-next-line react/forbid-dom-props */
                                style={{ backgroundColor: getLineStroke(line) }}
                                className={styles.legendMark}
                            />
                            {line.name}
                        </li>
                    ))}
                </ul>
            </div>
        </EventEmitterProvider>
    )
}
