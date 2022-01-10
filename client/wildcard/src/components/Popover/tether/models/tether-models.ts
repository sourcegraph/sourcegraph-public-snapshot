import { Rectangle } from './geometry/rectangle'

export enum Side {
    top = 'top',
    right = 'right',
    bottom = 'bottom',
    left = 'left',
}

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
    padding: Rectangle
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
     * Whenever tooltip doesn't have enough space pick any other position
     * whatever initial position tooltip has.
     *
     * Example: left → right → bottom → top
     */
    all = 'all',

    /**
     * Whenever tooltip doesn't have enough space pick only opposite position
     * of whatever initial position tooltip has.
     *
     * Example: left → right only
     * Example: top → bottom only
     */
    opposite = 'opposite',
}

/**
 * Describes all possible initial values for tooltip's position.
 */
export enum Position {
    topLeft = 'topLeft',
    topCenter = 'topCenter',
    topRight = 'topRight',
    leftTop = 'leftTop',
    leftMiddle = 'leftMiddle',
    leftBottom = 'leftBottom',
    rightTop = 'rightTop',
    rightMiddle = 'rightMiddle',
    rightBottom = 'rightBottom',
    bottomLeft = 'bottomLeft',
    bottomCenter = 'bottomCenter',
    bottomRight = 'bottomRight',
}
