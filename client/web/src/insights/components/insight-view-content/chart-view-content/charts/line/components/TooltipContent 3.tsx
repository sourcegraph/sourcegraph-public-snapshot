import { RenderTooltipParams } from '@visx/xychart/lib/components/Tooltip'
import classNames from 'classnames'
import React, { ReactElement } from 'react'
import { LineChartContent } from 'sourcegraph'

import { DEFAULT_LINE_STROKE } from '../constants'
import { Accessors } from '../types'

export interface TooltipContentProps<Datum extends object> extends RenderTooltipParams<Datum> {
    /** Accessors map to get information from nearest points. */
    accessors: Accessors<Datum, keyof Datum>
    /** Dataset of series (lines) on the chart. */
    series: LineChartContent<Datum, keyof Datum>['series']
    /** Possible className for root content element. */
    className?: string
}

/**
 * Display tooltip content for XYChart.
 * It consists of title - datetime for current x point and list of all nearest y points.
 */
export function TooltipContent<Datum extends object>(props: TooltipContentProps<Datum>): ReactElement | null {
    const { className = '', tooltipData, accessors, series } = props
    const datum = tooltipData?.nearestDatum?.datum

    if (!datum) {
        return null
    }

    const dateString = new Date(accessors.x(datum)).toDateString()
    const lineKeys = Object.keys(tooltipData?.datumByKey ?? {}).filter(lineKey => lineKey)

    return (
        <div className={classNames('line-chart__tooltip-content', className)}>
            <h3 className="line-chart__tooltip-date">{dateString}</h3>

            {/** values */}
            <ul className="line-chart__tooltip-list">
                {lineKeys.map(lineKey => {
                    const value = accessors.y[lineKey as keyof Datum](datum)
                    const line = series.find(line => line.dataKey === lineKey)
                    const datumKey = tooltipData?.nearestDatum?.key

                    /* eslint-disable react/forbid-dom-props */
                    return (
                        <li key={lineKey} className="line-chart__tooltip-item">
                            <em
                                className="line-chart__tooltip-item-name"
                                style={{
                                    color: line?.stroke ?? DEFAULT_LINE_STROKE,
                                    textDecoration: datumKey === lineKey ? 'underline' : undefined,
                                }}
                            >
                                {line?.name ?? 'unknown series'}
                            </em>{' '}
                            <span className="line-chart__tooltip-item-value">
                                {value === null || Number.isNaN(value) ? '–' : value}
                            </span>
                        </li>
                    )
                })}
            </ul>
        </div>
    )
}
