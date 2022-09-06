import { forwardRef, memo } from 'react'

import {
    AxisLeft as VisxAxisLeft,
    AxisBottom as VisxAsixBottom,
    TickLabelProps,
    SharedAxisProps,
    AxisScale,
} from '@visx/axis'
import { GridRows } from '@visx/grid'
import { Group } from '@visx/group'
import { TextProps } from '@visx/text'
import classNames from 'classnames'

import { Tick } from './Tick'
import { formatYTick, getXScaleTicks, getYScaleTicks } from './tick-formatters'

import styles from './Axis.module.scss'

// TODO: Remove this prop generation, see https://github.com/sourcegraph/sourcegraph/issues/39874
const getTickYLabelProps: TickLabelProps<number> = (value, index, values): Partial<TextProps> => ({
    dy: '0.25em',
    textAnchor: 'end',
    'aria-label': `Tick axis ${index + 1} of ${values.length}. Value: ${value}`,
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
                />

                <Group innerRef={reference} top={top} left={left}>
                    <VisxAxisLeft
                        {...attributes}
                        scale={scale}
                        tickValues={tickValues}
                        tickFormat={tickFormat}
                        tickLabelProps={getTickYLabelProps}
                        tickComponent={tickComponent}
                        axisLineClassName={classNames(styles.axisLine, styles.axisLineVertical)}
                        tickClassName={classNames(styles.axisTick, styles.axisTickVertical)}
                    />
                </Group>
            </>
        )
    })
)

AxisLeft.displayName = 'AxisLeft'

// TODO: Remove this prop generation, see https://github.com/sourcegraph/sourcegraph/issues/39874
const getTickXLabelProps: TickLabelProps<Date> = (value, index, values): Partial<TextProps> => ({
    'aria-label': `Tick axis ${index + 1} of ${values.length}. Value: ${value}`,
    textAnchor: 'middle',
})

export interface AxisBottomProps extends OwnSharedAxisProps {
    width: number
}

export const AxisBottom = memo(
    forwardRef<SVGGElement, AxisBottomProps>((props, reference) => {
        const { scale, top, left, width, tickValues, tickComponent = Tick, ...attributes } = props

        return (
            <Group innerRef={reference} top={top} left={left} width={width}>
                <VisxAsixBottom
                    {...attributes}
                    scale={scale}
                    tickComponent={tickComponent}
                    tickValues={tickValues ?? getXScaleTicks({ scale, space: width })}
                    tickLabelProps={getTickXLabelProps}
                    axisLineClassName={styles.axisLine}
                    tickClassName={styles.axisTick}
                />
            </Group>
        )
    })
)

AxisBottom.displayName = 'AxisBottom'
