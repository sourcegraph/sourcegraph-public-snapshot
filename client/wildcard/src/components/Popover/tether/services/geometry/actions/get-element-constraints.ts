import { createRectangleFromPoints, Rectangle } from '../../../models/geometry/rectangle';
import { Overlapping, Position, Side } from '../../../models/tether-models';
import { POSITION_VARIANTS } from '../constants';
import { createPoint } from '../../../models/geometry/point';

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
 * @param target - Target rectangle element
 * @param constraint - Original constraint rectangle element
 * @param position - Desirable tooltip position
 * @param overlapping - Overlapping strategy (all, none)
 */
export function getElementConstraint(
	target: Rectangle,
	constraint: Rectangle,
	position: Position,
	overlapping: Overlapping
): Rectangle {
	const side = POSITION_VARIANTS[position].positionSides;
	
	let x1 = constraint.left;
	let x2 = constraint.right;
	let y1 = constraint.top;
	let y2 = constraint.bottom;
	
	if (overlapping == Overlapping.none) {
		if (side == Side.top) {
			y1 = Math.min(target.top, constraint.top);
			y2 = Math.min(target.top, constraint.bottom);
		} else if (side == Side.right) {
			x1 = Math.max(target.right, constraint.left);
			x2 = Math.max(target.right, constraint.right);
		} else if (side == Side.bottom) {
			y1 = Math.max(target.bottom, constraint.top);
			y2 = Math.max(target.bottom, constraint.bottom);
		} else {
			x1 = Math.min(target.left, constraint.left);
			x2 = Math.min(target.left, constraint.right);
		}
	}
	
	return createRectangleFromPoints(
		createPoint(x1, y1),
		createPoint(x2, y2)
	);
}
