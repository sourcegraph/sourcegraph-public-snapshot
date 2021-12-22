import { ParentSize } from '@visx/responsive'
import { EventEmitterProvider } from '@visx/xychart'
import classNames from 'classnames'
import React, { ReactElement, useContext } from 'react'

import { getLineStroke, LineChartContent, LineChartContentProps } from './components/LineChartContent'
import { ScrollBox } from './components/scroll-box/ScrollBox'
import { MINIMAL_HORIZONTAL_LAYOUT_WIDTH, MINIMAL_SERIES_FOR_ASIDE_LEGEND } from './constants'
import { LineChartLayoutOrientation, LineChartSettingsContext } from './line-chart-settings-provider'
import styles from './LineChart.module.scss'

export interface LineChartProps<Datum extends object> extends LineChartContentProps<Datum> {
    /**
     * Whenever it is necessary to set size limits of line chart container block.
     * By default LineChart doesn't require
     */
    hasChartParentFixedSize?: boolean
}

/**
 * Display responsive line chart with legend below the chart.
 */
export function LineChart<Datum extends object>(props: LineChartProps<Datum>): ReactElement {
    const { width, height, hasChartParentFixedSize, ...otherProps } = props
    const { layout } = useContext(LineChartSettingsContext)

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
                style={hasChartParentFixedSize ? { width, height } : undefined}
                className={classNames(styles.lineChart, { [styles.lineChartHorizontal]: isHorizontal })}
            >
                {/*
                    In case if we have a legend to render we have to have responsive container for chart
                    just to calculate right sizes for chart content = rootContainerSizes - legendSizes
                */}
                <ParentSize className={styles.contentParentSize} data-line-chart-size-root="">
                    {({ width, height }) => <LineChartContent {...otherProps} width={width} height={height} />}
                </ParentSize>

                <ScrollBox
                    aria-hidden={true}
                    className={classNames(styles.legend, { [styles.legendHorizontal]: isHorizontal })}
                >
                    <ul className={classNames(styles.legendList, { [styles.legendListHorizontal]: isHorizontal })}>
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
                </ScrollBox>
            </div>
        </EventEmitterProvider>
    )
}
