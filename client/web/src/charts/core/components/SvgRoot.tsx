import {
    createContext,
    Dispatch,
    FC,
    PropsWithChildren, ReactNode,
    SetStateAction,
    useContext,
    useMemo,
    useState
} from 'react';

import { AxisScale, TickFormatter } from '@visx/axis/lib/types';
import { Group } from '@visx/group';
import { scaleLinear } from '@visx/scale';
import { noop } from 'lodash';
import useResizeObserver from 'use-resize-observer';

import { createRectangle, EMPTY_RECTANGLE, Rectangle } from '@sourcegraph/wildcard';

import { formatXTick } from '../../components/line-chart/utils';

import { AxisBottom, AxisLeft } from './axis/Axis';

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

interface SvgRootProps {
    width: number
    height: number
    yScale: AxisScale
    xScale: AxisScale
}

/**
 * SVG canvas root element. This component renders SVG element and
 * calculates and prepares all important canvas measurements for x/y axis,
 * content and other chart elements.
 */
export const SvgRoot: FC<PropsWithChildren<SvgRootProps>> = props => {
    const { width, height, yScale: yOriginalScale, xScale: xOriginalScale, children } = props

    const [padding, setPadding] = useState<Padding>(DEFAULT_PADDING)

    const contentRectangle = useMemo(
        () => createRectangle(
            padding.left,
            padding.top,
            width - padding.left - padding.right,
            height - padding.top - padding.bottom
        ) ,
        [width, height, padding]
    )

    const yScale = useMemo(
        () =>  yOriginalScale.copy().range([contentRectangle.height, 0]) as AxisScale,
        [yOriginalScale, contentRectangle]
    )

    const xScale = useMemo(
        () => xOriginalScale.copy().range([0, contentRectangle.width]) as AxisScale,
        [xOriginalScale, contentRectangle]
    )

    const context = useMemo<SVGRootLayout>(
        () => ({
            width,
            height,
            xScale,
            yScale,
            content: contentRectangle,
            setPadding
        }),
        [width, height, contentRectangle, xScale, yScale]
    )

    return (
        <SVGRootContext.Provider value={context}>
            <svg width={width} height={height} style={{ background: 'lightgoldenrodyellow', overflow: 'visible'}}>
                {children}
            </svg>
        </SVGRootContext.Provider>
    )
}

export const SvgAxisLeft: FC<{}> = props => {
    const { content, yScale, setPadding } = useContext(SVGRootContext)

    const handleResize = ({ width = 0 }): void => {
        setPadding(padding => ({ ...padding, left: width }))
    }

    const { ref } = useResizeObserver({ onResize: handleResize })

    return (
        <AxisLeft
            ref={ref}
            width={content.width}
            height={content.height}
            top={content.top}
            left={content.left}
            scale={yScale}
        />
    )
}

interface ObservedSize {
    width: number | undefined;
    height: number | undefined;
}

type ResizeHandler = (size: ObservedSize) => void;

export const SvgAxisBottom: FC<{rotate: boolean, tickFormat?: TickFormatter<AxisScale>}> = props => {
    const { content, xScale, setPadding } = useContext(SVGRootContext)

    const handleResize: ResizeHandler = ({ height = 0 }) => {
        setPadding(padding => ({ ...padding, bottom: height, }))
    }

    const { ref } = useResizeObserver<SVGGElement>({ onResize: handleResize })

    return (
        <AxisBottom
            ref={ref}
            rotate={props.rotate}
            scale={xScale}
            width={content.width}
            top={content.bottom}
            left={content.left}
            tickFormat={props.tickFormat ?? (formatXTick as unknown) as TickFormatter<AxisScale>}
        />
    )
}

interface SvgContentProps {
    children: (input: { yScale: AxisScale, xScale: AxisScale, content: Rectangle }) => ReactNode
}

export const SvgContent: FC<SvgContentProps> = props => {
    const { children } = props
    const { content, xScale, yScale, } = useContext(SVGRootContext)

    return (
        <Group
            top={content.top}
            left={content.left}
            width={content.width}
            height={content.height}>

            { children({ xScale, yScale, content }) }
        </Group>
    )
}
