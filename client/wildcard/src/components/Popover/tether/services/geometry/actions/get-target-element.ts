import { createPoint } from '../../../models/geometry/point'
import { createRectangleFromPoints, type Rectangle } from '../../../models/geometry/rectangle'
import type { ElementPosition } from '../../../models/tether-models'
import { POSITION_VARIANTS } from '../constants'

/**
 * Returns an extended target element rectangle by the marker rectangle size.
 *
 * Example: Extended target rect by marker size based on horizontal or
 * vertical tooltip position. Where x - marker width, y - marker height.
 * ```
 *                        ┌ ─ ─ ─ ───▲──
 *                         ░░░░░░░░│ │ y
 *               │  x │   ┣━━━━━━━━┳─▼──
 *               ◀────▶   ┃ Target ┃
 * ┌─ ──┏━━━━━━━━┫── ─┤   ┣━━━━━━━━┛
 *  ░░░░┃ Target ┃░░░░     ░░░░░░░░│
 * └─ ──┗━━━━━━━━┛── ─┘   └ ─ ─ ─ ─
 *```
 *
 * @param target - Original target rectangle.
 * @param marker - Marker (tail) rectangle.
 * @param position - Tooltip position relative to target element.
 */
export function getTargetElement(target: Rectangle, marker: Rectangle, position: ElementPosition): Rectangle {
    const offset = POSITION_VARIANTS[position].targetOffset

    const targetX = marker.width * offset.x
    const targetY = marker.height * offset.y

    // Extend width and height of tooltip element by marker rectangle
    const targetX1 = target.left - targetX
    const targetY1 = target.top - targetY
    const targetX2 = target.right + targetX
    const targetY2 = target.bottom + targetY

    return createRectangleFromPoints(createPoint(targetX1, targetY1), createPoint(targetX2, targetY2))
}
