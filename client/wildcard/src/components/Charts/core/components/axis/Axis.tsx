import { forwardRef, memo, useEffect } from 'react'

import {
    AxisLeft as VisxAxisLeft,
    AxisBottom as VisxAsixBottom,
    type TickLabelProps,
    type SharedAxisProps,
    type AxisScale,
} from '@visx/axis'
import { GridRows } from '@visx/grid'
import { Group } from '@visx/group'
import { useMergeRefs } from 'use-callback-ref'

import { Tick } from './Tick'
import { formatYTick, getXScaleTicks, getYScaleTicks } from './tick-formatters'

import styles from './Axis.module.scss'

// TODO: Remove this prop generation, see https://github.com/sourcegraph/sourcegraph/issues/39874
const getTickYLabelProps: TickLabelProps<number> = value => ({
    dy: '0.25em',
    textAnchor: 'end',
    'aria-label': `Axis tick, Value: ${value}`,
    role: 'listitem',
})

type OwnSharedAxisProps = Omit<SharedAxisProps<AxisScale>, 'tickLabelProps'>

export interface AxisLeftProps extends OwnSharedAxisProps {
    width: number
    height: number
    pixelsPerTick?: number
}

export const AxisLeft = memo(
    forwardRef<SVGGElement, AxisLeftProps>((props, reference) => {
        const {
            scale,
            left,
            top,
            width,
            height,
            pixelsPerTick = 40,
            tickComponent = Tick,
            tickFormat = formatYTick,
            tickValues = getYScaleTicks({ scale, space: height, pixelsPerTick }),
            ...attributes
        } = props

        return (
            <>
                <GridRows
                    top={top}
                    left={left}
                    width={width}
                    height={height}
                    scale={scale}
                    tickValues={tickValues}
                    className={styles.gridLine}
                    aria-hidden={true}
                />

                <Group innerRef={reference} top={top} left={left} role="list" aria-label="X axis">
                    <VisxAxisLeft
                        {...attributes}
                        scale={scale}
                        tickValues={tickValues}
                        tickFormat={tickFormat}
                        tickLabelProps={getTickYLabelProps}
                        tickComponent={tickComponent}
                        hideTicks={true}
                        hideAxisLine={true}
                    />
                </Group>
            </>
        )
    })
)

AxisLeft.displayName = 'AxisLeft'

// TODO: Remove this prop generation, see https://github.com/sourcegraph/sourcegraph/issues/39874
const getTickXLabelProps: TickLabelProps<Date> = value => ({
    'aria-label': `Axis tick, Value: ${value}`,
    textAnchor: 'middle',
    role: 'listitem',
})

export interface AxisBottomProps extends OwnSharedAxisProps {
    width: number
}

export const AxisBottom = memo(
    forwardRef<SVGGElement, AxisBottomProps>((props, reference) => {
        const { scale, top, left, width, tickValues, tickComponent = Tick, ...attributes } = props
        const mergedRef = useMergeRefs([reference])

        // Visx axis UI doesn't expose API to set attributes to axis line element
        // We need to hide axis line element in Safari from screen reader to avoid
        // extra "image" element announcement since this element is 100% decorative
        useEffect(() => {
            const axisGroup = mergedRef.current
            const axisLineElement = axisGroup?.querySelector(`.${styles.axisLine}`)

            axisLineElement?.setAttribute('aria-hidden', 'true')
        }, [mergedRef])

        return (
            <Group innerRef={mergedRef} top={top} left={left} width={width} role="list" aria-label="X axis">
                <VisxAsixBottom
                    {...attributes}
                    scale={scale}
                    tickComponent={tickComponent}
                    tickValues={tickValues ?? getXScaleTicks({ scale, space: width })}
                    tickLabelProps={getTickXLabelProps}
                    axisLineClassName={styles.axisLine}
                    tickClassName={styles.axisTick}
                    tickLineProps={{ 'aria-hidden': true }}
                />
            </Group>
        )
    })
)

AxisBottom.displayName = 'AxisBottom'
