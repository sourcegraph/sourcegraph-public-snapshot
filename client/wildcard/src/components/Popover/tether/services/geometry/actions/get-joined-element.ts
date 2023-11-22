import { createRectangle, type Rectangle } from '../../../models/geometry/rectangle'
import type { ElementPosition } from '../../../models/tether-models'
import { POSITION_VARIANTS } from '../constants'

/**
 * Returns joined tooltip element rectangle based on the target element position.
 *
 * Example: Return moved tooltip rectangle based on desirable tooltip position
 * and target position.
 *
 * ```
 * Top Left                         Right Top
 * ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
 *   ┌───────────┐                │   ┌───────────┐                │
 * │ │░░░░░░░░░░░│   New position   │ │░░░░░░░░░░░├─Shift
 *   │░░░░░░░░░░░│  ┌───────────┐ │   │░░░░░░░░░░░│  │             │
 * │ └─────┬─────┘  │░░░░░░░░░░░│   │ └───────────┘  │New position
 *         Shift───▶│░░░░░░░░░░░│ │          ┌──────┬▼──────────┐  │
 * │                ├──────┬────┘   │        │Target│░░░░░░░░░░░│
 *                  │Target│      │          └──────┤░░░░░░░░░░░│  │
 * │                └──────┘        │               └───────────┘
 *  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘
 *```
 *
 * @param element - Tooltip element rectangle
 * @param target - Target element rectangle
 * @param position - Desirable tooltip position
 */
export function getJoinedElement(element: Rectangle, target: Rectangle, position: ElementPosition): Rectangle {
    const elementA = POSITION_VARIANTS[position].elementAttachments
    const elementX = element.left + element.width * elementA.x
    const elementY = element.top + element.height * elementA.y

    const targetA = POSITION_VARIANTS[position].targetAttachments
    const targetX = target.left + target.width * targetA.x
    const targetY = target.top + target.height * targetA.y

    const shiftX = Math.floor(targetX - elementX)
    const shiftY = Math.floor(targetY - elementY)

    const pointX = element.left + shiftX
    const pointY = element.top + shiftY

    return createRectangle(pointX, pointY, element.width, element.height)
}
