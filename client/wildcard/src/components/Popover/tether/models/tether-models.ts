import type { Rectangle } from './geometry/rectangle'

/**
 * Describes padding (offset) of tooltip element and target element.
 */
export interface Padding {
    top: number
    left: number
    right: number
    bottom: number
}

/**
 * Describes boundary element for the tooltip position calculation.
 */
export interface Constraint {
    element: Rectangle
    padding: Padding
}

/**
 * Describes position strategy if tooltip doesn't have enough place to render
 */
export enum Overlapping {
    /**
     * Overlap all space of target element.
     */
    all = 'all',

    /**
     * Turns off position overlapping for any elements even if elements don't
     * have enough space.
     */
    none = 'none',
}

/**
 * Describes options for changing position when tooltip element doesn't have
 * enough space for rendering.
 */
export enum Flipping {
    /**
     * Whenever tooltip doesn't have enough space then pick any other position
     * based on initial position that tooltip has.
     *
     * Example: left → right → bottom → top
     */
    all = 'all',

    /**
     * Whenever tooltip doesn't have enough space then pick only opposite position
     * of whatever initial position tooltip got.
     *
     * Example: left → right only
     * Example: top → bottom only
     */
    opposite = 'opposite',
}

/**
 * Describes all possible initial values for tooltip position.
 */
export enum Position {
    topStart = 'topStart',
    top = 'top',
    topEnd = 'topEnd',
    leftStart = 'leftStart',
    left = 'left',
    leftEnd = 'leftEnd',
    rightStart = 'rightStart',
    right = 'right',
    rightEnd = 'rightEnd',
    bottomStart = 'bottomStart',
    bottom = 'bottom',
    bottomEnd = 'bottomEnd',
}

export type ElementPosition = Position | `${Position}`

export enum Strategy {
    /**
     * This strategy renders the element outside of DOM hierarchy in the designated
     * container in the body element and calculate position for fixed element.
     */
    Fixed = 'fixed',

    /**
     * Absolute strategy renders element right next to the target element and calculate
     * position based on the nearest container with position relative.
     */
    Absolute = 'absolute',
}
