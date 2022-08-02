import { forwardRef, memo, useMemo, useRef } from 'react'

import { AxisLeft as VisxAxisLeft, AxisBottom as VisxAsixBottom, TickLabelProps } from '@visx/axis'
import { AxisScale, TickRendererProps } from '@visx/axis/lib/types'
import { GridRows } from '@visx/grid'
import { Group } from '@visx/group'
import { TextProps } from '@visx/text'
import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref';

import { formatYTick, getXScaleTicks, getYScaleTicks } from '../../../components/line-chart/utils'

import { getMaxTickSize, Tick, TickProps } from './Tick'

import styles from './Axis.module.scss'

// TODO: Improve @visx/axis API in order to support value (tick) and values (ticks) props
// on component level and not just in prop function level
const getTickYLabelProps: TickLabelProps<number> = (value, index, values): Partial<TextProps> => ({
    dy: '0.25em',
    textAnchor: 'end',
    'aria-label': `Tick axis ${index + 1} of ${values.length}. Value: ${value}`,
})

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
                    innerRef={reference}
                    top={top}
                    left={left}
                >
                    <VisxAxisLeft
                        scale={scale}
                        tickValues={ticksValues}
                        tickFormat={formatYTick}
                        tickLabelProps={getTickYLabelProps}
                        tickComponent={Tick}
                        axisLineClassName={classNames(styles.axisLine, styles.axisLineVertical)}
                        tickClassName={classNames(styles.axisTick, styles.axisTickVertical)}
                    />
                </Group>
            </>
        )
    })
)

// TODO: Improve @visx/axis API in order to support value (tick) and values (ticks) props
// on component level and not just in prop function level
export const getTickXLabelProps: TickLabelProps<Date> = (value, index, values): Partial<TextProps> => ({
    'aria-label': `Tick axis ${index + 1} of ${values.length}. Value: ${value}`,
})

interface AxisBottomProps {
    rotate: boolean
    top: number
    left: number
    width: number
    scale: AxisScale
    tickFormat: (value: string) => string
}

export const AxisBottom = memo(
    forwardRef<SVGGElement, AxisBottomProps>((props, reference) => {
        const { scale, top, left, width, tickFormat } = props
        const groupReference = useRef<SVGGElement>(null)
        const ticks = getXScaleTicks({ scale, space: width, pixelsPerTick: 15 })
        const [, upperRangeBound] = scale.range() as [number, number]

        const maxWidth = useMemo(() => {
            if (!groupReference.current) {
                return 0
            }

            const { maxWidth } = getMaxTickSize(groupReference.current, ticks.map(tickFormat) )

            return maxWidth
        }, [ticks, tickFormat])

        const getXTickProps = (props: TickRendererProps): TickProps => {
            // TODO: Add more sophisticated logic around labels overlapping calculation
            const measuredSize = ticks.length * maxWidth;
            const rotate = upperRangeBound < measuredSize
                ? 90 * Math.min(1, (measuredSize / upperRangeBound - 0.8) / 2)
                : 0;

            if (rotate) {
                return {
                    ...props,
                    transform: `rotate(${-rotate}, ${props.x} ${props.y})`,
                    textAnchor: 'end',
                    maxWidth: 15
                }
            }

            return { ...props, textAnchor: 'middle' }
        };

        return (
            <Group innerRef={useMergeRefs([groupReference, reference])} top={top} left={left} width={width}>
                <VisxAsixBottom
                    scale={scale}
                    tickValues={ticks}
                    tickFormat={tickFormat}
                    tickLabelProps={getTickXLabelProps}
                    tickComponent={props => <Tick {...getXTickProps(props)}/>}
                    axisLineClassName={styles.axisLine}
                    tickClassName={styles.axisTick}
                />
            </Group>
        )
    })
)
