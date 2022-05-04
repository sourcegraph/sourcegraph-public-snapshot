import { FunctionComponent, useCallback } from 'react'

import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import { noop } from 'lodash'
import { ChartContent } from 'sourcegraph'

import { BarChart } from './charts/bar/BarChart'
import { LineChart } from './charts/line'
import { DatumZoneClickEvent } from './charts/line/types'
import { PieChart } from './charts/pie/PieChart'

import styles from './ChartViewContent.module.scss'

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
    content: ChartContent
    className?: string
    layout?: ChartViewContentLayout

    /**
     * Calls whenever the user clicks on or navigate through
     * the chart datum (pie arc, line point, bar category)
     */
    onDatumLinkClick?: () => void
    locked?: boolean
}

/**
 * Display chart content with different type of charts (line, bar, pie)
 */
export const ChartViewContent: FunctionComponent<React.PropsWithChildren<ChartViewContentProps>> = props => {
    const {
        content,
        className = '',
        layout = ChartViewContentLayout.ByParentSize,
        onDatumLinkClick = noop,
        locked = false,
    } = props

    // Click link-zone handler for line chart only. Catch click around point and redirect user by
    // link which we've got from the nearest datum point to user cursor position. This allows user
    // not to aim at small point on the chart and just click somewhere around the point.
    const linkHandler = useCallback(
        (event: DatumZoneClickEvent): void => {
            if (!event.link) {
                return
            }

            onDatumLinkClick()

            window.open(event.link, '_blank', 'noopener')?.focus()
        },
        [onDatumLinkClick]
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
                                onDatumLinkClick={onDatumLinkClick}
                                width={width}
                                height={height}
                                hasChartParentFixedSize={isResponsive}
                                locked={locked}
                            />
                        )
                    }

                    if (content.chart === 'bar') {
                        return (
                            <BarChart
                                {...content}
                                width={width}
                                height={height}
                                onDatumLinkClick={onDatumLinkClick}
                                locked={locked}
                            />
                        )
                    }

                    if (content.chart === 'pie') {
                        return (
                            <PieChart
                                {...content}
                                width={width}
                                height={height}
                                onDatumLinkClick={onDatumLinkClick}
                                locked={locked}
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
