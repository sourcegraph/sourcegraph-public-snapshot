import { createPoint, Point } from '../models/geometry/point'
import { createRectangle, createRectangleFromPoints, EMPTY_RECTANGLE, Rectangle } from '../models/geometry/rectangle'
import { Constraint, Flipping, Overlapping, Position, Strategy } from '../models/tether-models'

import { getAbsoluteAnchorOffset, getScrollParents } from './tether-browser'
import { Tether, TetherLayout } from './types'

/**
 * Collects all information about current layout (tether and popover elements rectangle),
 * constrains rectangles (viewport and scrollable parent elements).
 */
export function getLayout(tether: Tether): TetherLayout {
    const {
        position = Position.top,
        flipping = Flipping.opposite,
        overlapping = Overlapping.none,
        windowPadding = EMPTY_RECTANGLE,
        constraintPadding = EMPTY_RECTANGLE,
        overflowToScrollParents = true,
        strategy = Strategy.Fixed,
        constrainToScrollParents,
    } = tether

    const target = tether.pin
        ? createRectangleFromPoints(tether.pin, { x: tether.pin.x + 1, y: tether.pin.y + 1 })
        : tether.target
        ? tether.target.getBoundingClientRect()
        : EMPTY_RECTANGLE

    const element = tether.element.getBoundingClientRect()
    const marker = getMarkerSize(tether.marker)
    const anchorOffset = getAnchorOffset(tether.element, strategy)

    const overflows: Constraint[] = []
    const constraints: Constraint[] = []

    overflows.push({
        element: createRectangle(
            document.documentElement.clientLeft,
            document.documentElement.clientTop,
            document.documentElement.clientWidth,
            document.documentElement.clientHeight
        ),
        padding: windowPadding,
    })

    constraints.push({
        element: createRectangle(
            document.documentElement.clientLeft,
            document.documentElement.clientTop,
            document.documentElement.clientWidth,
            document.documentElement.clientHeight
        ),
        padding: windowPadding,
    })

    if (tether.constraint) {
        constraints.push({
            element: tether.constraint.getBoundingClientRect(),
            padding: constraintPadding,
        })
    }

    if (tether.target !== null && constrainToScrollParents) {
        const containers = getScrollParents(tether.target)

        const scrollConstraints = containers.map(container => ({
            element: container.getBoundingClientRect(),
            padding: constraintPadding,
        }))

        constraints.push(...scrollConstraints)
    }

    if (tether.target !== null && overflowToScrollParents && strategy === Strategy.Fixed) {
        const containers = getScrollParents(tether.target)

        const overflowConstraints = containers.map(container => ({
            element: container.getBoundingClientRect(),
            padding: EMPTY_RECTANGLE,
        }))

        overflows.push(...overflowConstraints)
    }

    return {
        element,
        target,
        marker,
        position,
        flipping,
        overlapping,
        overflows,
        constraints,
        anchorOffset,
        strategy,
    }
}

function getAnchorOffset(floating: HTMLElement, strategy: Strategy): Point {
    if (strategy === Strategy.Fixed) {
        return createPoint(0, 0)
    }

    return getAbsoluteAnchorOffset(floating)
}

function getMarkerSize(element?: HTMLElement | null): Rectangle {
    if (!element) {
        return createRectangle(0, 0, 0, 0)
    }

    // Reset transform rotation properties
    element.style.transform = ''

    // Measure element without rotation transformations
    // since transform rotation may affect element sizing
    return element.getBoundingClientRect()
}
