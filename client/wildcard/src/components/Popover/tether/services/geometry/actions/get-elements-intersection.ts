import { createPoint } from '../../../models/geometry/point'
import {
    createRectangle,
    createRectangleFromPoints,
    getIntersection,
    type Rectangle,
} from '../../../models/geometry/rectangle'
import type { Constraint, Padding } from '../../../models/tether-models'

import { getRoundedElement } from './rectangle-position-helpers'

/**
 * Applies constrains rectangles and returns intersection of all contained element.
 *
 * Example: Intersected rectangle based on two overlapped area.
 * ```
 *   ┌─────────────────┐
 * ┌─╋━━━━━━━━━━━━━━━━━╋─┐
 * │ ┃░░░░░░░░░░░░░░░░░┃ │
 * │ ┃░░░░░░░░░░░░░░░░░┃ │
 * └─╋━━━━━━━━━━━━━━━━━╋─┘
 *   └─────────────────┘
 * ```
 */
export function getElementsIntersection(constraints: Constraint[]): Rectangle {
    let constrainedArea: Rectangle | null = null

    for (const constraint of constraints) {
        const element = getRoundedElement(constraint.element)
        const content = getContentElement(element, constraint.padding)

        constrainedArea = constrainedArea ?? content
        constrainedArea = getIntersection(constrainedArea, content)
    }

    return constrainedArea ?? createRectangle(0, 0, 0, 0)
}

/**
 * Returns extended by padding rectangle.
 */
function getContentElement(element: Rectangle, padding: Padding): Rectangle {
    const xStart = element.left + padding.left
    const yStart = element.top + padding.top
    const xEnd = element.right - padding.right
    const yEnd = element.bottom - padding.bottom

    return createRectangleFromPoints(createPoint(xStart, yStart), createPoint(xEnd, yEnd))
}
