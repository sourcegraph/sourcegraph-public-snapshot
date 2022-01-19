import {
    getScrollPositions,
    isVisible,
    setMaxSize,
    setScrollPositions,
    setTransform,
    setVisibility,
} from './tether-browser'
import { getLayout } from './tether-layout'
import { getState } from './tether-state'
import { Tether } from './types'

/**
 * Main entry point for tooltip element position calculations. It mutates tooltip
 * DOM element (set visibility, width, height, top, left) properties.
 *
 * @param tether - settings tooltip information.
 * @param eventTarget - event target element event of which has triggered tooltip
 * position update.
 */
export function render(tether: Tether, eventTarget: HTMLElement | null): void {
    const positions = getScrollPositions(tether.element)

    if (!positions.points.has(eventTarget as Element)) {
        setMaxSize(tether.element, null)
    }

    // Restore visibility for correct measure in layout service
    setVisibility(tether.element, true)
    setVisibility(tether.marker ?? null, true)

    const layout = getLayout(tether)
    const state = getState(layout)

    if (state === null || !isVisible(tether.target)) {
        setVisibility(tether.element, false)
        setVisibility(tether.marker ?? null, false)

        return
    }

    setTransform(tether.element, 0, state.elementOffset)
    setTransform(tether.marker ?? null, state.markerAngle, state.markerOffset)

    if (!positions.points.has(eventTarget as Element)) {
        setMaxSize(tether.element, state.elementBounds)
        setScrollPositions(positions)
    }
}
