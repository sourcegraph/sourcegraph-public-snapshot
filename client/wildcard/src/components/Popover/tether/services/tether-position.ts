import { createPoint, type Point } from '../models/geometry/point'
import { getIntersection, intersects, type Rectangle } from '../models/geometry/rectangle'
import { type ElementPosition, Strategy } from '../models/tether-models'

import {
    getConstrainedElement,
    getElementBounds,
    getElementConstraint,
    getElementsIntersection,
    getJoinedElement,
    getMarkerConstraint,
    getMarkerOffset,
    getMarkerRotation,
    getRoundedElement,
    getTargetElement,
    isElementVisible,
} from './geometry'
import { getExtendedConstraint } from './geometry/actions/get-extended-constraint'
import type { TetherLayout } from './types'

export interface TetherState {
    /** Area of the element in pixels */
    elementArea: number

    /** Y and X coordinates of tooltip element */
    elementOffset: Point

    /** Constrained tooltip element due to constraints */
    elementBounds: Rectangle | null

    /** Tooltip tail angle based on tooltip element position */
    markerAngle: number

    /** Y and X coordinates of marker element */
    markerOffset: Point

    position: ElementPosition
}

/**
 * Calculates position for the tooltip element based on layout settings and current position value
 * of the tooltip element. In case if there is no visual way to fit tooltip element in the constraints
 * then returns null.
 *
 * @param layout - Document layout information (overflows, constrains, paddings, etc)
 * @param position - Another position value to fit tooltip element
 */
export function getPositionState(layout: TetherLayout, position: ElementPosition): TetherState | null {
    const { overlapping, anchorOffset, strategy, targetPadding } = layout
    const { element, target, marker, overflow, constraint } = getNormalizedLayout(layout)

    const { markerAngle, markerOrigin, rotatedMarker } = getMarkerRotation(marker, position)

    // Apply overflow constraints to target element
    const overflowedTarget = strategy === Strategy.Fixed ? getIntersection(target, overflow) : target

    // Force tooltip layout hide in case if target is outside of overflow constraint.
    if (!isElementVisible(overflowedTarget)) {
        return null
    }

    // Extend the target element by marker size element  for correctness of calculations below
    const targetElement = getTargetElement(
        getTargetElement(overflowedTarget, targetPadding, position),
        rotatedMarker,
        position
    )

    // Change element tooltip coordinates to put this element right next extended target element
    const joinedElement = getJoinedElement(element, targetElement, position)

    const extendedConstraint =
        strategy === Strategy.Absolute ? getExtendedConstraint(targetElement, constraint) : constraint

    // Calculate constraint rectangle by target position and default constraint
    const elementConstraint = getElementConstraint(targetElement, extendedConstraint, position, overlapping)

    // Change tooltip element rectangle by element constraint
    const constrainedElement = getConstrainedElement(joinedElement, elementConstraint)

    // Calculate element metric of the tooltip element after all constraint calculations
    const elementArea = constrainedElement.width * constrainedElement.height
    const elementOffset = createPoint(constrainedElement.left + anchorOffset.x, constrainedElement.top + anchorOffset.y)
    const elementBounds = getElementBounds(constrainedElement, element)

    // Change element tooltip coordinates to put the element right next target element
    const joinedMarker = getJoinedElement(rotatedMarker, overflowedTarget, position)

    // Calculate constraint rectangle for marker element (always within tooltip element rectangle)
    const markerConstraint = getMarkerConstraint(constrainedElement, joinedMarker, position)

    // Apply marker constraint (shift marker element if it's needed)
    const constrainedMarker = getConstrainedElement(joinedMarker, markerConstraint)
    const markerOffset = getMarkerOffset(markerOrigin, constrainedMarker)

    // Check visibility and join geometry
    const isTooltipVisible = isElementVisible(constrainedElement)
    const isTooltipJoined = intersects(targetElement, constrainedElement)
    const isMarkerJoined = intersects(targetElement, constrainedMarker)

    if (!isTooltipVisible || (!isTooltipJoined && !isMarkerJoined)) {
        return null
    }

    return {
        elementArea,
        elementOffset,
        elementBounds,
        markerAngle,
        markerOffset,
        position,
    }
}

// ------------ Private methods --------------

interface NormalizedTetherLayout {
    element: Rectangle
    target: Rectangle
    marker: Rectangle
    overflow: Rectangle
    constraint: Rectangle
}

/**
 * Returns normalized (rounded and amplified) layout input.
 */
function getNormalizedLayout(layout: TetherLayout): NormalizedTetherLayout {
    const { element, target, marker, overflows, constraints } = layout

    return {
        element: getRoundedElement(element),
        target: getRoundedElement(target),
        marker: getRoundedElement(marker),
        overflow: getElementsIntersection(overflows),
        constraint: getElementsIntersection(constraints),
    }
}
