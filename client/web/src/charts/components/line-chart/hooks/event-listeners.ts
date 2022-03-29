import { MouseEvent, MouseEventHandler, PointerEventHandler } from 'react'

import { localPoint } from '@visx/event'
import { Point } from '@visx/point'

interface UseChartEventHandlersProps {
    onPointerMove: (point: Point) => void
    onPointerLeave: () => void
    onClick: MouseEventHandler<SVGSVGElement>
}

interface ChartHandlers {
    onPointerMove: PointerEventHandler<SVGSVGElement>
    onPointerLeave: PointerEventHandler<SVGSVGElement>
    onClick: MouseEventHandler<SVGSVGElement>
}

/**
 * Provides special svg|chart-specific handlers for mouse/touch events.
 */
export function useChartEventHandlers(props: UseChartEventHandlersProps): ChartHandlers {
    const { onPointerMove, onPointerLeave, onClick } = props

    const handleMouseMove: MouseEventHandler<SVGGElement> = event => {
        const point = localPoint(event.currentTarget, event)

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
        onPointerMove: handleMouseMove,
        onPointerLeave: handleMouseOut,
        onClick,
    }
}
