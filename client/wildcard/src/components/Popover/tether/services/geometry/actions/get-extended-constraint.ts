import { createPoint } from '../../../models/geometry/point'
import { createRectangleFromPoints, type Rectangle } from '../../../models/geometry/rectangle'

/**
 * Extends constraint rectangle by target rectangle position.
 *
 * ```
 *          ┏━━━━━━━━┓      ┌ ── ── ┳━━━━━━━━┓
 *          ┃ Target ┃      │░░░░░░░┃ Target ┃
 *  ┌───────┫        ┃       ░░░░░░░┃        ┃
 *  │░░░░░░░┗━━━━┳━━━┛      │░░░░░░░┗━━━━━━━━┫
 *  │░░░░░░░░░░░░│ ───────▶ │░░░░░░░░░░░░░░░░│
 *  │░░░░░░░░░░░░│           ░░░░░░░░░░░░░░░░
 *  │░░░░░░░░░░░░│          │░░░░░░░░░░░░░░░░│
 *  └Constraint──┘          └Constraint── ── ┘
 * ```
 */
export function getExtendedConstraint(target: Rectangle, constraint: Rectangle): Rectangle {
    const xStart = Math.min(constraint.left, Math.max(target.left, target.right))
    const yStart = Math.min(constraint.top, Math.max(target.top, target.bottom))
    const xEnd = Math.max(constraint.right, Math.min(target.left, target.right))
    const yEnd = Math.max(constraint.bottom, Math.min(target.top, target.bottom))

    return createRectangleFromPoints(createPoint(xStart, yStart), createPoint(xEnd, yEnd))
}
