import { createPoint } from '../../../models/geometry/point'
import { createRectangleFromPoints, type Rectangle } from '../../../models/geometry/rectangle'
import { type ElementPosition, Overlapping, Position } from '../../../models/tether-models'
import { POSITION_VARIANTS } from '../constants'

/**
 * Returns constraint rectangle according to target element position
 * tooltip element position and overlapping setting.
 *
 * Example: Pattern area for right, left, bottom or bottom tooltip positioned
 * constraint. If overlap all just returns original constraint.
 * ```
 * ┌────────┳━━━━━━━━━┓ ┏━━━━━━━━━┳────────┐
 * │        ┃░░░░░░░░░┃ ┃░░░░░░░░░┃        │
 * │┌──────┐┃░░░░░░░░░┃ ┃░░░░░░░░░┃┌──────┐│
 * ││Target│┃░░░░░░░░░┃ ┃░░░░░░░░░┃│Target││
 * │└──────┘┃░░░░░░░░░┃ ┃░░░░░░░░░┃└──────┘│
 * │        ┃░░░░░░░░░┃ ┃░░░░░░░░░┃        │
 * └────────┻━━━━━━━━━┛ ┗━━━━━━━━━┻────────┘
 * ┏━━━━━━━━━━━━━━━━━━┓ ┌──────────────────┐
 * ┃░░░░░░░░░░░░░░░░░░┃ │     ┌──────┐     │
 * ┃░░░░░░░░░░░░░░░░░░┃ │     │Target│     │
 * ┣━━━━━━━━━━━━━━━━━━┫ │     └──────┘     │
 * │     ┌──────┐     │ ┣━━━━━━━━━━━━━━━━━━┫
 * │     │Target│     │ ┃░░░░░░░░░░░░░░░░░░┃
 * │     └──────┘     │ ┃░░░░░░░░░░░░░░░░░░┃
 * └──────────────────┘ ┗━━━━━━━━━━━━━━━━━━┛
 * ```
 *
 * @param target - Target rectangle element
 * @param constraint - Original constraint rectangle element
 * @param position - Desirable tooltip position
 * @param overlapping - Overlapping strategy (all, none)
 */
export function getElementConstraint(
    target: Rectangle,
    constraint: Rectangle,
    position: ElementPosition,
    overlapping: Overlapping
): Rectangle {
    const side = POSITION_VARIANTS[position].positionSides

    let xStart = constraint.left
    let xEnd = constraint.right
    let yStart = constraint.top
    let yEnd = constraint.bottom

    if (overlapping === Overlapping.none) {
        if (side === Position.top) {
            yStart = Math.min(target.top, constraint.top)
            yEnd = Math.min(target.top, constraint.bottom)
        } else if (side === Position.right) {
            xStart = Math.max(target.right, constraint.left)
            xEnd = Math.max(target.right, constraint.right)
        } else if (side === Position.bottom) {
            yStart = Math.max(target.bottom, constraint.top)
            yEnd = Math.max(target.bottom, constraint.bottom)
        } else {
            xStart = Math.min(target.left, constraint.left)
            xEnd = Math.min(target.left, constraint.right)
        }
    }

    return createRectangleFromPoints(createPoint(xStart, yStart), createPoint(xEnd, yEnd))
}
