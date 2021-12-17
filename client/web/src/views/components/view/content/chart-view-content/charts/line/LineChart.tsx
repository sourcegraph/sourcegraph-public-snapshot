import { ParentSize } from '@visx/responsive'
import { EventEmitterProvider } from '@visx/xychart'
import classNames from 'classnames'
import React, { ReactElement, useContext } from 'react'

import { getLineStroke, LineChartContent, LineChartContentProps } from './components/LineChartContent'
import { ScrollBox } from './components/scroll-box/ScrollBox'
import { MINIMAL_HORIZONTAL_LAYOUT_WIDTH, MINIMAL_SERIES_FOR_ASIDE_LEGEND } from './constants'
import { LineChartLayoutOrientation, LineChartSettingsContext } from './line-chart-settings-provider'
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

    const hasViewManySeries = otherProps.series.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
    const hasEnoughXSpace = width >= MINIMAL_HORIZONTAL_LAYOUT_WIDTH

    const isHorizontal = layout
        ? // If layout is defined explicitly in line chart setting context use its value
          layout === LineChartLayoutOrientation.Horizontal
        : // Otherwise apply internal logic (based on how many x space and series we have)
          hasViewManySeries && hasEnoughXSpace

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

                <ScrollBox
                    as="ul"
                    scrollEnabled={isHorizontal}
                    aria-hidden={true}
                    rootClassName={classNames({ [styles.legendHorizontal]: isHorizontal })}
                    className={classNames(styles.legendContent, { [styles.legendContentHorizontal]: isHorizontal })}
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
                </ScrollBox>
            </div>
        </EventEmitterProvider>
    )
}
