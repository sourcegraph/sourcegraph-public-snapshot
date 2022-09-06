import {
    createContext,
    Dispatch,
    FC,
    PropsWithChildren,
    ReactElement,
    ReactNode,
    SetStateAction,
    SVGProps,
    useContext,
    useMemo,
    useRef,
    useState,
} from 'react'

import { AxisScale, TickRendererProps } from '@visx/axis'
import { Group } from '@visx/group'
import { scaleLinear } from '@visx/scale'
import { noop } from 'lodash'
import { useMergeRefs } from 'use-callback-ref'
import useResizeObserver from 'use-resize-observer'

// In order to resolve cyclic deps in tests
// see https://github.com/sourcegraph/sourcegraph/pull/40209#pullrequestreview-1069334480
import { createRectangle, EMPTY_RECTANGLE, Rectangle } from '../../../Popover'

import { AxisBottom, AxisLeft } from './axis/Axis'
import { getMaxTickWidth, Tick, TickProps } from './axis/Tick'
import { GetScaleTicksOptions, getXScaleTicks } from './axis/tick-formatters'

const DEFAULT_PADDING = { top: 16, right: 36, bottom: 0, left: 0 }

interface Padding {
    top: number
    right: number
    bottom: number
    left: number
}

interface SVGRootLayout {
    width: number
    height: number
    yScale: AxisScale
    xScale: AxisScale
    content: Rectangle
    setPadding: Dispatch<SetStateAction<Padding>>
}

const SVGRootContext = createContext<SVGRootLayout>({
    width: 0,
    height: 0,
    xScale: scaleLinear(),
    yScale: scaleLinear(),
    content: EMPTY_RECTANGLE,
    setPadding: noop,
})

interface SvgRootProps extends SVGProps<SVGSVGElement> {
    width: number
    height: number
    yScale: AxisScale
    xScale: AxisScale
}

/**
 * SVG canvas root element. This component renders SVG element and
 * calculates and prepares all important canvas measurements for x/y-axis,
 * content and other chart elements.
 */
export const SvgRoot: FC<PropsWithChildren<SvgRootProps>> = props => {
    const { width, height, yScale: yOriginalScale, xScale: xOriginalScale, children, ...attributes } = props

    const [padding, setPadding] = useState<Padding>(DEFAULT_PADDING)

    const contentRectangle = useMemo(
        () =>
            createRectangle(
                padding.left,
                padding.top,
                width - padding.left - padding.right,
                height - padding.top - padding.bottom
            ),
        [width, height, padding]
    )

    const yScale = useMemo(() => yOriginalScale.copy().range([contentRectangle.height, 0]) as AxisScale, [
        yOriginalScale,
        contentRectangle,
    ])

    const xScale = useMemo(() => xOriginalScale.copy().range([0, contentRectangle.width]) as AxisScale, [
        xOriginalScale,
        contentRectangle,
    ])

    const context = useMemo<SVGRootLayout>(
        () => ({
            width,
            height,
            xScale,
            yScale,
            content: contentRectangle,
            setPadding,
        }),
        [width, height, contentRectangle, xScale, yScale]
    )

    return (
        <SVGRootContext.Provider value={context}>
            <svg {...attributes} width={width} height={height}>
                {children}
            </svg>
        </SVGRootContext.Provider>
    )
}

interface SvgAxisLeftProps {
    pixelsPerTick?: number
}

export const SvgAxisLeft: FC<SvgAxisLeftProps> = props => {
    const { content, yScale, setPadding } = useContext(SVGRootContext)

    const handleResize = ({ width = 0 }): void => {
        setPadding(padding => ({ ...padding, left: width }))
    }

    const { ref } = useResizeObserver({ onResize: handleResize })

    return (
        <AxisLeft
            {...props}
            ref={ref}
            width={content.width}
            height={content.height}
            top={content.top}
            left={content.left}
            scale={yScale}
        />
    )
}

const defaultToString = <T,>(tick: T): string => `${tick}`
const defaultTruncatedTick = (tick: string): string => (tick.length >= 15 ? `${tick.slice(0, 15)}...` : tick)

// TODO: Support reverse truncation for some charts https://github.com/sourcegraph/sourcegraph/issues/39879
export const reverseTruncatedTick = (tick: string): string => (tick.length >= 15 ? `...${tick.slice(-15)}` : tick)

interface SvgAxisBottomProps<Tick> {
    tickFormat?: (tick: Tick) => string
    pixelsPerTick?: number
    minRotateAngle?: number
    maxRotateAngle?: number
    hideTicks?: boolean
    getTruncatedTick?: (formattedTick: string) => string
    getScaleTicks?: <T>(options: GetScaleTicksOptions) => T[]
}
export function SvgAxisBottom<Tick = string>(props: SvgAxisBottomProps<Tick>): ReactElement {
    const {
        pixelsPerTick = 0,
        minRotateAngle = 0,
        maxRotateAngle = 45,
        tickFormat = defaultToString,
        getTruncatedTick = defaultTruncatedTick,
        getScaleTicks = getXScaleTicks,
        hideTicks = false,
    } = props
    const { content, xScale, setPadding } = useContext(SVGRootContext)

    const axisGroupRef = useRef<SVGGElement>(null)
    const { ref } = useResizeObserver<SVGGElement>({
        // TODO: Fix corner cases with axis sizes see https://github.com/sourcegraph/sourcegraph/issues/39876
        onResize: ({ height = 0 }) => setPadding(padding => ({ ...padding, bottom: height })),
    })

    const [, upperRangeBound] = xScale.range() as [number, number]
    const ticks = getScaleTicks<Tick>({ scale: xScale, space: content.width, pixelsPerTick })

    const maxWidth = useMemo(() => {
        const axisGroup = axisGroupRef.current

        if (!axisGroup) {
            return 0
        }

        return getMaxTickWidth(axisGroup, ticks.map(tickFormat))
    }, [tickFormat, ticks])

    const getXTickProps = (props: TickRendererProps): TickProps => {
        // TODO: Improve rotation math see https://github.com/sourcegraph/sourcegraph/issues/41310
        const measuredSize = ticks.length * maxWidth
        const fontSize = 12 // 0.75rem
        const rotate =
            upperRangeBound < measuredSize
                ? Math.max(maxRotateAngle * Math.min(1, (measuredSize / upperRangeBound - 0.8) / 2), minRotateAngle)
                : 0

        if (rotate) {
            return {
                ...props,
                // Truncate ticks only if we rotate them, this means truncate labels only
                // when they overlap
                getTruncatedTick,
                transform: `rotate(${rotate}, ${props.x + fontSize / 2} ${props.y - fontSize / 2})`,
                textAnchor: 'start',
            }
        }

        return { ...props, textAnchor: 'middle' }
    }

    return (
        <AxisBottom
            ref={useMergeRefs([axisGroupRef, ref])}
            scale={xScale}
            width={content.width}
            top={content.bottom}
            left={content.left}
            tickValues={ticks}
            tickComponent={props => <Tick {...getXTickProps(props)} />}
            tickFormat={tickFormat}
            hideTicks={hideTicks}
        />
    )
}

interface SvgContentProps<XScale extends AxisScale, YScale extends AxisScale> {
    children: (input: { xScale: XScale; yScale: YScale; content: Rectangle }) => ReactNode
}

/**
 * Compound svg canvas component, to render actual chart content on
 * SVG canvas with pre-calculated axes and paddings
 */
export function SvgContent<XScale extends AxisScale = AxisScale, YScale extends AxisScale = AxisScale>(
    props: SvgContentProps<XScale, YScale>
): ReactElement {
    const { children } = props
    const { content, xScale, yScale } = useContext(SVGRootContext)

    return (
        <Group top={content.top} left={content.left} width={content.width} height={content.height}>
            {children({
                // We need to cast scales here because there is no other way to type context
                // shared data in TS, React interfaces.
                xScale: xScale as XScale,
                yScale: yScale as YScale,
                content,
            })}
        </Group>
    )
}
