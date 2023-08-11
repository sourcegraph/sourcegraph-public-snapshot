import type { Point } from './point'

export interface Rectangle {
    left: number
    top: number
    right: number
    bottom: number
    width: number
    height: number
}

export const createRectangle = (left: number, top: number, rawWidth: number, rawHeight: number): Rectangle => {
    // Inlined clamp to zero logic.
    const width = rawWidth < 0 ? 0 : rawWidth
    const height = rawHeight < 0 ? 0 : rawHeight

    return {
        left,
        top,
        width,
        height,

        // Additional calculated properties
        right: left + width,
        bottom: top + height,
    }
}

export const EMPTY_RECTANGLE = createRectangle(0, 0, 0, 0)

/**
 * Create rectangle from two points by subtracting one point from other.
 */
export const createRectangleFromPoints = (a: Point, b: Point): Rectangle => {
    const left = Math.min(a.x, b.x)
    const top = Math.min(a.y, b.y)
    const width = Math.max(a.x, b.x) - left
    const height = Math.max(a.y, b.y) - top

    return createRectangle(left, top, width, height)
}

/**
 * Returns intersection of two rectangles. Returns empty rectangle in case
 * if input rectangles don't have any intersection area.
 */
export function getIntersection(a: Rectangle, b: Rectangle): Rectangle {
    const xStart = Math.max(a.left, b.left)
    const xEnd = Math.min(a.left + a.width, b.left + b.width)

    if (xStart <= xEnd) {
        const yStart = Math.max(a.top, b.top)
        const yEnd = Math.min(a.top + a.height, b.top + b.height)

        if (yStart <= yEnd) {
            return createRectangle(xStart, yStart, xEnd - xStart, yEnd - yStart)
        }
    }

    return createRectangle(0, 0, 0, 0)
}

/**
 * Tests that two rectangles do or don't have overlapping between each other.
 */
export function intersects(a: Rectangle, b: Rectangle): boolean {
    return (
        a.left <= b.left + b.width &&
        b.left <= a.left + a.width &&
        a.top <= b.top + b.height &&
        b.top <= a.top + a.height
    )
}

/**
 * Tests does the first rectangle contain the second one.
 */
export function containsRectangle(a: Rectangle, b: Rectangle): boolean {
    return (
        a.left <= b.left &&
        a.left + a.width >= b.left + b.width &&
        a.top <= b.top &&
        a.top + a.height >= b.top + b.height
    )
}
