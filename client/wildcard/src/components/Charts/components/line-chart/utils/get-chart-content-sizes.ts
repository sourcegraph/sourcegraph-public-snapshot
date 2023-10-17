import type { Optional } from 'utility-types'

// In order to resolve cyclic deps in tests
// see https://github.com/sourcegraph/sourcegraph/pull/40209#pullrequestreview-1069334480
import { createRectangle, type Rectangle } from '../../../../Popover'

interface GetChartContentSizesInput {
    width: number
    height: number
    margin?: Optional<Margin>
}

interface Margin {
    top: number
    right: number
    bottom: number
    left: number
}

/**
 * It applies margin values to the original container rectangle and calculate
 * content width and height values.
 * ```
 *  Original size container
 * ┌──────────▲─────────┐
 * │░░░░░░░top│░░░░░░░░░│
 * │░░░░░┌────────┐░░░░░│
 * │left░│        │right│
 * ◀─────│ Content├─────▶
 * │░░░░░│        │░░░░░│
 * │░░░░░└────┬───┘░░░░░│
 * │░░░░░░░░░░│bottom░░░│
 * └──────────▼─────────┘
 * ```
 */
export function getChartContentSizes(input: GetChartContentSizesInput): Rectangle {
    const { width, height, margin = {} } = input

    const { top, left, bottom, right } = {
        top: Math.round(margin.top ?? 0),
        right: Math.round(margin.right ?? 0),
        bottom: Math.round(margin.bottom ?? 0),
        left: Math.round(margin.left ?? 0),
    }

    return createRectangle(left, top, width - left - right, height - top - bottom)
}
