import { createPoint, type Point } from '../../../models/geometry/point'
import { createRectangle, createRectangleFromPoints, type Rectangle } from '../../../models/geometry/rectangle'
import { type ElementPosition, Position } from '../../../models/tether-models'
import { POSITION_VARIANTS } from '../constants'

/**
 * Returns rectangle (constraint) that marker should always be within.
 *
 * @param element - constrained tooltip element rectangle
 * @param marker - rotated marker
 * @param position - another tooltip position
 */
export function getMarkerConstraint(element: Rectangle, marker: Rectangle, position: ElementPosition): Rectangle {
    const side = POSITION_VARIANTS[position].positionSides

    let xStart = element.right
    let xEnd = element.left
    let yStart = element.bottom
    let yEnd = element.top

    const deltaX = Math.floor(Math.min(marker.width, Math.floor((element.width - marker.width) / 2)))
    const deltaY = Math.floor(Math.min(marker.height, Math.floor((element.height - marker.height) / 2)))

    if (side === Position.top) {
        xStart = element.left + deltaX
        xEnd = element.right - deltaX
        yEnd = element.bottom + marker.height
    } else if (side === Position.right) {
        xStart = element.left - marker.width
        yStart = element.top + deltaY
        yEnd = element.bottom - deltaY
    } else if (side === Position.bottom) {
        xStart = element.left + deltaX
        xEnd = element.right - deltaX
        yStart = element.top - marker.height
    } else {
        xEnd = element.right + marker.width
        yStart = element.top + deltaY
        yEnd = element.bottom - deltaY
    }

    return createRectangleFromPoints(createPoint(xStart, yStart), createPoint(xEnd, yEnd))
}

export function getMarkerOffset(origin: Point, marker: Rectangle): Point {
    return createPoint(marker.left + origin.x, marker.top + origin.y)
}

interface MarkerRotation {
    markerOrigin: Point
    markerAngle: number
    rotatedMarker: Rectangle
}

/**
 * Returns marker element position information. Shifted center of marker element for
 * correct rotation, marker rectangle itself and rotation angle.
 */
export function getMarkerRotation(marker: Rectangle, position: ElementPosition): MarkerRotation {
    const markerAngle = POSITION_VARIANTS[position].rotationAngle

    if (markerAngle % 180 !== 0) {
        const delta = Math.floor((marker.width - marker.height) / 2)
        const markerOrigin = createPoint(-delta, delta)

        // Swap marker height and width and therefore rotate.
        const rotatedMarker = createRectangle(marker.left, marker.top, marker.height, marker.width)

        return {
            rotatedMarker,
            markerOrigin,
            markerAngle,
        }
    }
    return {
        markerOrigin: createPoint(0, 0),
        rotatedMarker: marker,
        markerAngle,
    }
}
