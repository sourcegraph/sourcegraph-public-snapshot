import { createRectangle, createRectangleFromPoints, Rectangle } from '../../../models/geometry/rectangle';
import { Position, Side } from '../../../models/tether-models';
import { POSITION_VARIANTS } from '../constants';
import { createPoint, Point } from '../../../models/geometry/point';

/**
 * Returns rectangle (constraint) that marker should always be within.
 *
 * @param element - constrained tooltip element rectangle
 * @param marker - rotated marker
 * @param position - another tooltip position
 */
export function getMarkerConstraint(element: Rectangle, marker: Rectangle, position: Position): Rectangle {
	const side = POSITION_VARIANTS[position].positionSides;
	
	let x1 = element.right;
	let x2 = element.left;
	let y1 = element.bottom;
	let y2 = element.top;
	
	const deltaX = Math.floor(Math.min(marker.width, Math.floor((element.width - marker.width) / 2)));
	const deltaY = Math.floor(Math.min(marker.height, Math.floor((element.height - marker.height) / 2)));
	
	if (side == Side.top) {
		x1 = element.left + deltaX;
		x2 = element.right - deltaX;
		y2 = element.bottom + marker.height;
	} else if (side == Side.right) {
		x1 = element.left - marker.width;
		y1 = element.top + deltaY;
		y2 = element.bottom - deltaY;
	} else if (side == Side.bottom) {
		x1 = element.left + deltaX;
		x2 = element.right - deltaX;
		y1 = element.top - marker.height;
	} else {
		x2 = element.right + marker.width;
		y1 = element.top + deltaY;
		y2 = element.bottom - deltaY;
	}
	
	return createRectangleFromPoints(
		createPoint(x1, y1),
		createPoint(x2, y2)
	);
}

export function getMarkerOffset(origin: Point, marker: Rectangle): Point {
	return createPoint(marker.left + origin.x, marker.top + origin.y);
}

interface MarkerRotation {
	markerOrigin: Point,
	markerAngle: number
	rotatedMarker: Rectangle,
}

/**
 * Returns marker element position information. Shifted center of marker element for
 * correct rotation, marker rectangle itself and rotation angle.
 */
export function getMarkerRotation(marker: Rectangle, position: Position): MarkerRotation {
	const markerAngle = POSITION_VARIANTS[position].rotationAngle;
	
	if (markerAngle % 180 != 0) {
		const delta = Math.floor((marker.width - marker.height) / 2);
		const markerOrigin = createPoint(-delta, delta);
		
		// Swap marker height and width and therefore rotate.
		const rotatedMarker = createRectangle(marker.left, marker.top, marker.height, marker.width)
		
		return {
			rotatedMarker,
			markerOrigin,
			markerAngle
		};
	} else {
		return {
			markerOrigin: createPoint(0, 0),
			rotatedMarker: marker,
			markerAngle
		};
	}
}
