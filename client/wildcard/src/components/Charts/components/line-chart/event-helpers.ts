import { localPoint } from '@visx/event'
import type { VoronoiDiagram, VoronoiSite } from 'd3-voronoi'

interface Offset {
    top: number
    left: number
}

export function getClosesVoronoiPoint<T>(
    event: PointerEvent,
    voronoiLayout: VoronoiDiagram<T>,
    // Taking into account content area shift in point distribution map
    // see https://github.com/sourcegraph/sourcegraph/issues/38919
    offset: Offset
): VoronoiSite<T> | null {
    const point = localPoint(event.currentTarget as Element, event)

    return point && voronoiLayout.find(point.x - offset.left, point.y - offset.top)
}

export function isNextTargetWithinCurrent(event: PointerEvent | FocusEvent, excludeRoot = true): boolean {
    const relatedTarget = event.relatedTarget as Element
    const currentTarget = event.currentTarget as Element

    // Contains on the root element (like svg.contains(svg)) always
    // returns true, to avoid it handle root case manually.
    if (excludeRoot && currentTarget === relatedTarget) {
        return false
    }

    return currentTarget.contains(relatedTarget)
}
