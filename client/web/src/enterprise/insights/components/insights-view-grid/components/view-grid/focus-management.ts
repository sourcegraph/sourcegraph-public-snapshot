import type { Layout } from 'react-grid-layout'

export type Direction = 'top' | 'right' | 'bottom' | 'left'

/**
 * It calculates the next element into the grid based on the looking direction and the
 * current position. It implements combined approach of two - wrapped list navigation logic
 * and 2d elements grid.
 */
export function findNextLayout(direction: Direction, anchorElementId: string, layout: Layout[]): Layout | null {
    const currentElement = layout.find(element => element.i === anchorElementId)

    if (!currentElement) {
        return null
    }

    switch (direction) {
        case 'top': {
            return findTopElement(layout, currentElement)
        }
        case 'bottom': {
            return findBottomElement(layout, currentElement)
        }
        case 'left': {
            const leftElement = findClosesLeftElement(layout, currentElement)

            if (!leftElement) {
                const firstElement = findFirstElement(layout)
                const isFirstElementInRow = currentElement.i === firstElement?.i

                if (isFirstElementInRow) {
                    return findLastElement(layout)
                }

                return findTopRightElement(layout, currentElement)
            }

            return leftElement
        }
        case 'right': {
            const rightElement = findClosesRightElement(layout, currentElement)

            if (!rightElement) {
                const firstElement = findLastElement(layout)
                const isFirstElementInRow = currentElement.i === firstElement?.i

                if (isFirstElementInRow) {
                    return findFirstElement(layout)
                }

                return findBottomLeftElement(layout, currentElement)
            }

            return rightElement
        }
    }
}

/**
 * Searches for the top element relative to the current one.
 * ```
 * ┌────┐ ┌────┐ ┌────┐
 * │    │ │    │◀┼┐   │
 * └────┘ └────┘ └┼───┘
 * ┌────┐ ┌────┐  │
 * │    │ │    │──┘
 * └────┘ └────┘
 * ```
 */
function findTopElement(layout: Layout[], currentElement: Layout): Layout | null {
    const getBottomCoordinate = (element: Layout): number => element.y + element.h

    const topElements = layout
        .filter(element => getBottomCoordinate(element) <= currentElement.y)
        .sort((element1, element2) => getBottomCoordinate(element2) - getBottomCoordinate(element1))
        .reduce((elements, element) => {
            if (elements.length === 0) {
                return [element]
            }

            const prevElement = elements[0]

            if (getBottomCoordinate(element) === getBottomCoordinate(prevElement)) {
                elements.push(element)
            }

            return elements
        }, [] as Layout[])

    return topElements.reduce((element1, element2) => {
        if (!element1) {
            return element2
        }

        const magnitude1 = getMagnitude(element1, currentElement)
        const magnitude2 = getMagnitude(element2, currentElement)

        if (magnitude1 === magnitude2 && element1.x < element2.x) {
            return element1
        }

        if (magnitude1 >= magnitude2) {
            return element2
        }

        return element1
    }, null as Layout | null)
}

/**
 * Searches for the bottom element relative to the current one.
 * ```
 * ┌────┐ ┌────┐ ┌────┐
 * │    │ │    │─┼┐   │
 * └────┘ └────┘ └┼───┘
 * ┌────┐ ┌────┐  │
 * │    │ │    │◀─┘
 * └────┘ └────┘
 * ```
 */
function findBottomElement(layout: Layout[], currentElement: Layout): Layout | null {
    const nextLayouts = layout
        .filter(element => element.y >= currentElement.y + currentElement.h)
        .sort((element1, element2) => Math.abs(currentElement.x - element1.x) - Math.abs(currentElement.x - element2.x))
        .sort((element1, element2) => element1.y - element2.y)

    return nextLayouts[0] ?? null
}

/**
 * It looks for the closest left element relative to the target in the grid layout.
 * ```
 * ┌────┐ ┌────┐ ┌────┐
 * │    │ │    │ │    │
 * └────┘ │    │ └────┘
 * ┌────┐ │    │
 * │    │ └────┘
 * └────┘ ┌────┐
 *    ▲┌──│    │
 *    ││  └────┘
 *    └┘
 * ```
 */
function findClosesLeftElement(layout: Layout[], currentElement: Layout): Layout | null {
    if (layout.length === 0) {
        return null
    }

    const leftElements = layout.filter(element => element.x < currentElement.x && element.y === currentElement.y)

    return leftElements.reduce((element1, element2) => {
        if (!element1) {
            return element2
        }

        const magnitude1 = getMagnitude(element1, currentElement)
        const magnitude2 = getMagnitude(element2, currentElement)

        const isElementsInOneRow = element1.y === element2.y
        const isElement2CloserByX = magnitude2 < magnitude1

        if (isElementsInOneRow && isElement2CloserByX) {
            return element2
        }

        if (magnitude1 > magnitude2) {
            return element2
        }

        return element1
    }, null as Layout | null)
}

/**
 * It searches for the top right element (the last element in the previous row) relative to the
 * current element. (wrapped list navigation)
 *
 * ```
 * ┌────┐ ┌────┐ ┌────┐
 * │    │ │    │ │    │◀─┐
 * └────┘ │    │ └────┘  │
 * ┌────┐ │    │         │
 * │    │ └────┘         │
 * └────┘ ┌────┐         │
 *    │   │    │         │
 *    │   └────┘         │
 *    └──────────────────┘
 * ```
 */
function findTopRightElement(layout: Layout[], currentElement: Layout): Layout | null {
    if (layout.length === 0) {
        return null
    }

    const topElements = layout
        .filter(element => element.y < currentElement.y)
        .sort((elem1, elem2) => elem2.y - elem1.y)
        .reduce((elements, elem) => {
            if (elements.length === 0) {
                return [elem]
            }

            if (elements[0].y === elem.y) {
                elements.push(elem)
            }

            return elements
        }, [] as Layout[])
        .sort((elem1, elem2) => elem2.x - elem1.x)

    return topElements[0] ?? null
}

/**
 * It searches for the bottom left element (the first element in the next row) relative to the
 * current element. (wrapped list navigation)
 * ```
 * ┌────────────────────────┐
 * │                        │
 * │  ┌────┐ ┌────┐ ┌────┐  │
 * │  │    │ │    │ │    │──┘
 * │  └────┘ │    │ └────┘
 * │  ┌────┐ │    │
 * └─▶│    │ └────┘
 *    └────┘ ┌────┐
 *           │    │
 *           └────┘
 * ```
 */
function findBottomLeftElement(layout: Layout[], currentElement: Layout): Layout | null {
    if (layout.length === 0) {
        return null
    }

    const bottomElements = layout
        .filter(element => element.y > currentElement.y)
        .sort((elem1, elem2) => elem1.y - elem2.y)
        .reduce((elements, elem) => {
            if (elements.length === 0) {
                return [elem]
            }

            if (elements[0].y === elem.y) {
                elements.push(elem)
            }

            return elements
        }, [] as Layout[])
        .sort((elem1, elem2) => elem1.x - elem2.x)

    return bottomElements[0] ?? null
}

/**
 * It looks for the closest left element relative to the target in the grid layout.
 *
 * ```
 *    ┌────┐ ┌────┐ ┌────┐
 *    │    │ │    │ │    │
 *    └────┘ │    │ └────┘
 *    ┌────┐ │    │
 * ┌──│    │ └────┘
 * │  └────┘ ┌────┐
 * └────────▶│    │
 *           └────┘
 * ```
 */
function findClosesRightElement(layout: Layout[], currentElement: Layout): Layout | null {
    if (layout.length === 0) {
        return null
    }

    const currentX = currentElement.x + currentElement.w
    const currentY = currentElement.y

    const rightElements = layout.filter(element => element.x >= currentX && element.y >= currentY)

    return rightElements.reduce((element1, element2) => {
        if (!element1) {
            return element2
        }

        const magnitude1 = getMagnitude(element1, currentElement)
        const magnitude2 = getMagnitude(element2, currentElement)

        const isElementsInOneRow = element1.y === element2.y
        const isElement2CloserByX = magnitude2 < magnitude1

        if (isElementsInOneRow && isElement2CloserByX) {
            return element2
        }

        if (magnitude1 > magnitude2) {
            return element2
        }

        return element1
    }, null as Layout | null)
}

function findFirstElement(layout: Layout[]): Layout | null {
    if (layout.length === 0) {
        return null
    }

    return layout.reduce((element1, element2) => {
        if (element1 === null) {
            return element2
        }

        if (element2.y < element1.y && element2.x < element1.x) {
            return element2
        }

        return element1
    }, null as Layout | null)
}

function findLastElement(layout: Layout[]): Layout | null {
    if (layout.length === 0) {
        return null
    }

    return layout.reduce((element1, element2) => {
        if (element1 === null) {
            return element2
        }

        if (element2.y > element1.y) {
            return element2
        }

        if (element2.y === element1.y) {
            return element2.x > element1.x ? element2 : element1
        }

        return element1
    }, null as Layout | null)
}

interface Point {
    x: number
    y: number
}

function getMagnitude(point1: Point, point2: Point): number {
    const distanceY = Math.abs(point1.y - point2.y)
    const distanceX = Math.abs(point1.x - point2.x)

    return Math.sqrt(Math.pow(distanceY, 2) + Math.pow(distanceX, 2))
}
