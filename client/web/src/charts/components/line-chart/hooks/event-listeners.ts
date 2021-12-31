import { localPoint } from '@visx/event'
import { Point } from '@visx/point'
import { MouseEvent, MouseEventHandler, PointerEventHandler, RefObject, useRef } from 'react'

interface UseChartEventHandlersProps {
    onPointerMove: (point: Point) => void
    onPointerLeave: () => void
    onClick: () => void
}

interface ChartHandlers {
    root: RefObject<SVGSVGElement>
    onPointerMove: PointerEventHandler<SVGSVGElement>
    onPointerLeave: PointerEventHandler<SVGSVGElement>
    onClick: MouseEventHandler<SVGSVGElement>
}

export function useChartEventHandlers(props: UseChartEventHandlersProps): ChartHandlers {
    const { onPointerMove, onPointerLeave, onClick } = props
    const rootSvgReference = useRef<SVGSVGElement>(null)

    const handleMouseMove: MouseEventHandler<SVGGElement> = event => {
        if (!rootSvgReference.current) {
            return
        }

        const point = localPoint(rootSvgReference.current, event)

        if (!point) {
            return
        }

        onPointerMove(point)
    }

    const handleMouseOut = (event: MouseEvent): void => {
        let relatedTarget = event.relatedTarget as Element

        while (relatedTarget) {
            // go up the parent chain and check – if we're still inside currentElem
            // then that's an internal transition – ignore it
            if (relatedTarget === event.currentTarget) {
                return
            }

            relatedTarget = relatedTarget?.parentNode as Element
        }

        onPointerLeave()
    }

    return {
        root: rootSvgReference,
        onPointerMove: handleMouseMove,
        onPointerLeave: handleMouseOut,
        onClick,
    }
}
