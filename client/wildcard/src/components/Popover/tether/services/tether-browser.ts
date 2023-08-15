import { createPoint, type Point } from '../models/geometry/point'
import type { Rectangle } from '../models/geometry/rectangle'
import type { ElementPosition } from '../models/tether-models'

import { POSITION_VARIANTS } from './geometry/constants'

interface ScrollPositions {
    points: WeakMap<Element, Point>
    elements: Element[]
}

/**
 * Collect elements with scroll overflow (with scrollbar) and returns
 * their scroll position coordinates (scrollX, scrollY)
 *
 * @param element - the root element for getting scroll information
 */
export function getScrollPositions(element: Element): ScrollPositions {
    const store = new WeakMap<Element, Point>()
    const scrollElements = getScrollChildren(element)

    for (const element of scrollElements) {
        store.set(element, createPoint(element.scrollLeft, element.scrollTop))
    }

    return { elements: scrollElements, points: store }
}

/**
 * Returns list of all parent elements that have scroll.
 */
export function getScrollParents(element: Element): HTMLElement[] {
    const containers: HTMLElement[] = []

    if (element.parentElement !== null) {
        if (isScrollContainer(element.parentElement)) {
            containers.push(element.parentElement)
        }

        containers.push(...getScrollParents(element.parentElement))
    }

    return containers
}

/**
 * Iterates over elements and restore they scroll positions by information
 * that was collected by @link getScrollPositions function.
 *
 * @param positions - scroll positions information.
 */
export function setScrollPositions(positions: ScrollPositions): void {
    for (const container of positions.elements) {
        const position = positions.points.get(container)

        container.scrollLeft = position?.x ?? 0
        container.scrollTop = position?.y ?? 0
    }
}

/**
 * Walks DOM from the element to top-level HTML-elements (body, html) and
 * checks visibility properties in order to be sure that the element is visually
 * visible.
 */
export function isVisible(element: HTMLElement | null): boolean {
    let current = element

    while (current) {
        if (current.hidden !== null && current.hidden) {
            return false
        }

        current = current.parentElement
    }

    return true
}

/**
 * Returns offset of the element that is parent of the target element and at the same time
 * creates another stacking context.
 *
 * ```
 *   ┌────▲────────────────────────┐
 *   │    │y                       │
 *   ◀────╋━━stacking context━━━┓  │
 *   │ x  ┃  ┌── ─── ─── ──┐    ┃  │
 *   │    ┃   ┌ ─ ─ ─ ─ ─ ┐     ┃  │
 *   │    ┃  │  ┌──────┐   │    ┃  │
 *   │    ┃  ││ │Target│  ││    ┃  │
 *   │    ┃  │  └──────┘   │    ┃  │
 *   │    ┃   └ ─ ─ ─ ─ ─ ┘     ┃  │
 *   │    ┃  └── ─── ─── ──┘    ┃  │
 *   │    ┗━━━━━━━━━━━━━━━━━━━━━┛  │
 *   └─────────────────────────────┘
 * ```
 */
export function getAbsoluteAnchorOffset(element: HTMLElement): Point {
    let current = element.parentElement

    while (current) {
        const styles = getComputedStyle(current)

        if (styles.position !== 'static') {
            const rectangle = current.getBoundingClientRect()

            return createPoint(current.scrollLeft - rectangle.left, current.scrollTop - rectangle.top)
        }

        current = current.parentElement
    }

    return createPoint(0, 0)
}

export function setTransform(element: HTMLElement | SVGElement | null, angle: number, offset: Point): void {
    setStyle(element, 'transform', `translate(${offset.x}px, ${offset.y}px) rotate(${angle}deg)`)
}

export function setMaxSize(element: HTMLElement, bounds: Rectangle | null): void {
    setStyle(element, 'max-width', bounds !== null ? `${bounds.width}px` : '')
    setStyle(element, 'max-height', bounds !== null ? `${bounds.height}px` : '')
}

export function setVisibility(element: HTMLElement | null, isVisible: boolean): void {
    if (element !== null && element.hidden !== !isVisible) {
        element.hidden = !isVisible
        element.style.setProperty('visibility', isVisible ? 'visible' : 'hidden')
    }
}

export function setStyle(element: HTMLElement | SVGElement | null, key: string, value: string): void {
    if (element !== null && element.style.getPropertyValue(key) !== value) {
        element.style.setProperty(key, value)
    }
}

export function setPositionAttributes(element: HTMLElement | SVGElement | null, position: ElementPosition): void {
    if (element !== null && position) {
        element.dataset.position = position
        element.dataset.side = POSITION_VARIANTS[position].positionSides
    }
}

// ------------- Private API methods ---------------

/**
 * Collect all elements by the root element and below that have scroll.
 */
function getScrollChildren(element: Element): Element[] {
    const containers: Element[] = []

    if (isScrollContainer(element)) {
        containers.push(element)
    }

    for (const child of [...element.children]) {
        containers.push(...getScrollChildren(child))
    }

    return containers
}

function isScrollContainer(element: Element): boolean {
    if (element.scrollWidth > element.clientWidth || element.scrollHeight > element.clientHeight) {
        const style = getComputedStyle(element)
        const keywords = new Set(['auto', 'scroll', 'overlay'])
        const properties = [style.overflow, style.overflowX, style.overflowY]

        return properties.some(property => keywords.has(property))
    }

    return false
}
