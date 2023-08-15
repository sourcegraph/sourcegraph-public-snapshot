import { createPoint } from '../../../models/geometry/point'
import { createRectangleFromPoints, type Rectangle } from '../../../models/geometry/rectangle'

/**
 * Moves element rectangle in a way to fit the element rectangle to the constraint element.
 * If moving isn't possible cuts off some parts of overflowed tooltip element rectangle.
 *
 * ```
 *         ┏━━━━━━━━┓      New position
 * ┌───────╋░░░░░░░░┃     ┌───┳━━━━━━━━┓
 * │       ┃░░░░░░░░┃────┐│   ┃░░░░░░░░┃
 * │       ┗━━━━╋━━━┛    └┼──▶┃░░░░░░░░┃
 * │            │         │   ┗━━━━━━━━┫
 * │            │         │            │
 * └────────────┘         └────────────┘
 * ```
 *
 * @param element - Tooltip rectangle element
 * @param constraint - Constraint rectangle.
 */
export function getConstrainedElement(element: Rectangle, constraint: Rectangle): Rectangle {
    const xStart = Math.max(Math.min(element.left, constraint.right - element.width), constraint.left)
    const yStart = Math.max(Math.min(element.top, constraint.bottom - element.height), constraint.top)
    const xEnd = Math.min(Math.max(element.right, constraint.left + element.width), constraint.right)
    const yEnd = Math.min(Math.max(element.bottom, constraint.top + element.height), constraint.bottom)

    return createRectangleFromPoints(createPoint(xStart, yStart), createPoint(xEnd, yEnd))
}
