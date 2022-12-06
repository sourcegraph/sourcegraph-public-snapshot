import { localPoint } from '@visx/event'
import { VoronoiDiagram, VoronoiSite } from 'd3-voronoi'

interface Offset {
    top: number
    left: number
}

export function getClosesVoronoiPoint<T>(
    event: PointerEvent,
    voronoiLayout: VoronoiDiagram<T>,
    offset: Offset
): VoronoiSite<T> | null {
    const point = localPoint(event.currentTarget as Element, event)

    return point && voronoiLayout.find(point.x - offset.left, point.y - offset.top)
}

export function isNextTargetWithinCurrent(event: PointerEvent | FocusEvent): boolean {
    const relatedTarget = event.relatedTarget as Element
    const currentTarget = event.currentTarget as Element

    return currentTarget.contains(relatedTarget)
}
