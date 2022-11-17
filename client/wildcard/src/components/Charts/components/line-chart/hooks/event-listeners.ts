import { FocusEventHandler, MouseEvent, MouseEventHandler, PointerEventHandler } from 'react'

import { localPoint } from '@visx/event'
import { Point } from '@visx/point'

interface UseChartEventHandlersProps {
    onPointerMove: (point: Point) => void
    onPointerLeave: () => void
    onFocusOut: FocusEventHandler<SVGGraphicsElement>
    onClick: MouseEventHandler<SVGGraphicsElement>
}

interface ChartHandlers {
    onPointerMove: PointerEventHandler<SVGGraphicsElement>
    onPointerLeave: PointerEventHandler<SVGGraphicsElement>
    onBlurCapture: FocusEventHandler<SVGGraphicsElement>
    onClick: MouseEventHandler<SVGGraphicsElement>
}

/**
 * Provides special svg|chart-specific handlers for mouse/touch events.
 */
export function useChartEventHandlers(props: UseChartEventHandlersProps): ChartHandlers {
    const { onPointerMove, onPointerLeave, onFocusOut, onClick } = props

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

    const handleBlurCapture: FocusEventHandler<SVGSVGElement> = event => {
        const relatedTarget = event.relatedTarget as Element
        const currentTarget = event.currentTarget as Element

        if (!currentTarget.contains(relatedTarget)) {
            onFocusOut(event)
        }
    }

    return {
        onPointerMove: handleMouseMove,
        onPointerLeave: handleMouseOut,
        onBlurCapture: handleBlurCapture,
        onClick,
    }
}
