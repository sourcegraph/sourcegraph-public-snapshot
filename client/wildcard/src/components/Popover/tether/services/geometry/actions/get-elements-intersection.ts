import { Constraint, Padding } from '../../../models/tether-models';
import {
	createRectangle,
	createRectangleFromPoints,
	intersection,
	Rectangle
} from '../../../models/geometry/rectangle';
import { createPoint } from '../../../models/geometry/point';
import { getRoundedElement } from './rectangle-position-helpers';

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
export function getElementsIntersection(constraints: Array<Constraint>): Rectangle {
	let constrainedArea: Rectangle | null = null;
	
	for (const constraint of constraints) {
		const element = getRoundedElement(constraint.element);
		const content = getContentElement(element, constraint.padding);
		
		constrainedArea = constrainedArea ?? content;
		constrainedArea = intersection(constrainedArea, content);
	}
	
	return constrainedArea ?? createRectangle(0, 0, 0, 0);
}

/**
 * Returns extended by padding rectangle.
 */
function getContentElement(element: Rectangle, padding: Padding): Rectangle {
	const x1 = element.left + padding.left;
	const y1 = element.top + padding.top;
	const x2 = element.right - padding.right;
	const y2 = element.bottom - padding.bottom;
	
	return createRectangleFromPoints(
		createPoint(x1, y1),
		createPoint(x2, y2)
	);
}
