import type { Point } from '../models/geometry/point'
import type { Rectangle } from '../models/geometry/rectangle'
import type { Constraint, ElementPosition, Flipping, Overlapping, Padding, Strategy } from '../models/tether-models'

export interface Tether {
    /** Reference on target HTML element in the DOM. */
    target: HTMLElement | null

    /** Reference on tooltip HTML element in the DOM. */
    element: HTMLElement

    /** Reference on tooltip tail (marker) HTML element in the DOM. */
    marker?: MarkerElement | null

    /**
     * In case if consumer wants to position not by target but
     * just a point on the page. It could be mouse cursor or
     * some coordinate on a canvas chart.
     */
    pin?: Point | null

    /**
     * Initial tooltip position. It can be changed by tooltip position calculator
     * during position calculation that takes into account layout data (constraints,
     * viewport space, paddings, etc)
     */
    position?: ElementPosition

    /**
     * Position flipping strategy. With active flipping tooltip tries to find other
     * position on opposite side of target element if current doesn't have enough
     * space.
     */
    flipping?: Flipping

    /**
     * Position strategy settings, allows tooltip to overlap target element if it's
     * needed (not enough space with current position)
     */
    overlapping?: Overlapping

    /**
     * A custom constrain element for the tooltip element position.
     */
    constraint?: HTMLElement

    /**
     * Setups position strategy (Fixed or Absolute) to render the popover element.
     */
    strategy?: Strategy

    targetPadding?: Rectangle
    windowPadding?: Partial<Padding>
    constraintPadding?: Partial<Padding>

    overflowToScrollParents?: boolean
    constrainToScrollParents?: boolean
}

export type MarkerElement = HTMLElement | SVGElement

export interface TetherLayout {
    /** Tooltip target element */
    target: Rectangle

    /** Tooltip HTML element itself */
    element: Rectangle

    /** Marker HTML element (for tooltip tail/arrow rendering) */
    marker: Rectangle

    /** Tooltip position relative to target */
    position: ElementPosition

    /** Position flipping strategy */
    flipping: Flipping

    /** Overlapping position strategy */
    overlapping: Overlapping

    overflows: Constraint[]
    constraints: Constraint[]

    strategy: Strategy
    anchorOffset: Point
    targetPadding: Rectangle
}
