import { forwardRef, memo } from 'react'

import { AxisLeft as VisxAxisLeft, AxisBottom as VisxAsixBottom } from '@visx/axis'
import { AxisScale, TickFormatter } from '@visx/axis/lib/types'
import { GridRows } from '@visx/grid'
import { Group } from '@visx/group'
import classNames from 'classnames'

import { formatYTick, getXScaleTicks, getYScaleTicks } from '../../../components/line-chart/utils'

import { getTickXProps, getTickYProps, Tick } from './Tick'

import styles from './Axis.module.scss'

interface AxisLeftProps {
    top: number
    left: number
    width: number
    height: number
    scale: AxisScale
}

export const AxisLeft = memo(
    forwardRef<SVGGElement, AxisLeftProps>((props, reference) => {
        const { scale, left, top, width, height } = props
        const ticksValues = getYScaleTicks({ scale, space: height })

        return (
            <>
                <GridRows
                    top={top}
                    left={left}
                    width={width}
                    height={height}
                    scale={scale}
                    tickValues={ticksValues}
                    className={styles.gridLine}
                />

                <Group
                    key={ticksValues.reduce((store, tick) => `${store}-${tick}`, '')}
                    innerRef={reference}
                    top={top}
                    left={left}
                >
                    <VisxAxisLeft
                        scale={scale}
                        tickValues={ticksValues}
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
)

interface AxisBottomProps {
    top: number
    left: number
    width: number
    scale: AxisScale
    tickFormat?: TickFormatter<AxisScale>
}

export const AxisBottom = memo(
    forwardRef<SVGGElement, AxisBottomProps>((props, reference) => {
        const { scale, top, left, width, tickFormat } = props

        return (
            <Group innerRef={reference} top={top} left={left}>
                <VisxAsixBottom
                    scale={scale}
                    tickValues={getXScaleTicks({ scale, space: width })}
                    tickFormat={tickFormat}
                    tickLabelProps={getTickXProps}
                    tickComponent={Tick}
                    axisLineClassName={styles.axisLine}
                    tickClassName={styles.axisTick}
                />
            </Group>
        )
    })
)
