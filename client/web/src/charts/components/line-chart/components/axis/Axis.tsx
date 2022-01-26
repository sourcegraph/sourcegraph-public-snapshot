import { AxisLeft as VisxAxisLeft, AxisBottom as VisxAsixBottom } from '@visx/axis'
import { AxisScale } from '@visx/axis/lib/types'
import { GridRows } from '@visx/grid'
import { Group } from '@visx/group'
import classNames from 'classnames'
import React, { forwardRef } from 'react'

import { formatXTick, formatYTick, getXScaleTicks, getYScaleTicks } from '../../utils/ticks'

import styles from './Axis.module.scss'
import { getTickXProps, getTickYProps, Tick } from './Tick'

interface AxisLeftProps {
    top: number
    left: number
    width: number
    height: number
    scale: AxisScale
}

export const AxisLeft = forwardRef<SVGGElement, AxisLeftProps>((props, reference) => {
    const { scale, left, top, width, height } = props

    return (
        <>
            <GridRows
                top={top}
                left={left}
                width={width - left}
                height={height}
                scale={scale}
                tickValues={getYScaleTicks({ scale, space: height })}
                className={styles.gridLine}
            />

            <Group innerRef={reference} top={top} left={left}>
                <VisxAxisLeft
                    scale={scale}
                    tickValues={getYScaleTicks({ scale, space: height })}
                    tickFormat={formatYTick}
                    tickLabelProps={getTickYProps}
                    tickComponent={Tick}
                    axisLineClassName={classNames(styles.axisLine, styles.axisLineVertical)}
                    tickClassName={classNames(styles.axisTick, styles.axisTickVertical)}
                />
            </Group>
        </>
    )
})

interface AxisBottomProps {
    top: number
    width: number
    scale: AxisScale
}

export const AxisBottom = forwardRef<SVGGElement, AxisBottomProps>((props, reference) => {
    const { scale, top, width } = props

    return (
        <Group innerRef={reference} top={top}>
            <VisxAsixBottom
                scale={scale}
                tickValues={getXScaleTicks({ scale, space: width })}
                tickFormat={formatXTick}
                tickLabelProps={getTickXProps}
                tickComponent={Tick}
                axisLineClassName={styles.axisLine}
                tickClassName={styles.axisTick}
            />
        </Group>
    )
})
