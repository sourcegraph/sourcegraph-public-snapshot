
export interface Point {
	x: number;
	y: number;
}

export const createPoint = (x: number, y: number): Point => ({ x, y })

/**
 * Compare two points, returns true if points have the same coordinates value.
 */
export const isPointsEqual = (point1: Point | undefined | null, point2: Point| undefined | null): boolean =>
	!!point1 && !!point2 && point1.x === point2.x && point1.y === point2.y;

